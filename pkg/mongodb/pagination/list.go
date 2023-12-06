package pagination

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

const (
	DEFAULT_PAGE_SIZE = 20
)

type Paginator interface {
	AddPipelineStages(stages []bson.D, resultColumns map[string]Columned, hasChangeableResultCount bool)
	AddPipelineStage(stage bson.D, resultColumns map[string]Columned, isChangeableResultCount bool)
	RunQuery(ctx context.Context, fromCollection *mongo.Collection, fromCollectionColumnsMap map[string]Columned, defaultSort string, dataSliceType interface{})
}

type List struct {
	Filters         map[string]string `json:"filters"`
	Sort            string            `json:"sort"`
	PageNo          int               `json:"pageNo"`
	PageSize        int               `json:"pageSize"`
	TotalItemsCount int               `json:"totalItemsCount"`

	data                     interface{}
	dbPipelines              []pipeline
	sortStageAdded           bool
	fromCollection           *mongo.Collection
	fromCollectionColumnsMap map[string]Columned
}

type pipeline struct {
	Pipeline              bson.D
	ChangeableResultCount bool // ChangeableResultCount this field is true when stage is match, group or another stages that can change result count
	ResultColumns         map[string]Columned
}

// New create pagination pagination
func New() Paginator {
	return &List{
		PageSize: DEFAULT_PAGE_SIZE,
		PageNo:   1,
	}
}

// AddPipelineStages add multiple stages to pipeline and ChangeableResultCount this field is true when stage is match, group or another filter stages
func (list *List) AddPipelineStages(stages []bson.D, resultColumns map[string]Columned, hasChangeableResultCount bool) {
	if len(stages) > 1 {
		for _, stage := range stages[:len(stages)-1] {
			list.AddPipelineStage(stage, nil, false)
		}
	}
	/// add resultColumns after all pipelines
	if len(stages) > 0 {
		list.dbPipelines = append(list.dbPipelines, pipeline{
			Pipeline:              stages[len(stages)-1],
			ResultColumns:         resultColumns,
			ChangeableResultCount: hasChangeableResultCount,
		})
	}
}

// AddPipelineStage add multiple stages to pipeline and ChangeableResultCount this field is true when stage is match, group or another filter stages
func (list *List) AddPipelineStage(stage bson.D, resultColumns map[string]Columned, isChangeableResultCount bool) {
	list.dbPipelines = append(list.dbPipelines, pipeline{
		Pipeline:              stage,
		ResultColumns:         resultColumns,
		ChangeableResultCount: isChangeableResultCount,
	})
}

// RunQuery run pipeline stages and filter by pagination filters
func (list *List) RunQuery(ctx context.Context, fromCollection *mongo.Collection, fromCollectionColumnsMap map[string]Columned, defaultSort string, dataSliceType any) {
	list.fromCollection = fromCollection
	list.fromCollectionColumnsMap = fromCollectionColumnsMap
	if list.Sort == "" {
		list.Sort = defaultSort
	}
	dbPipelines, countPipelines := list.mergePaginationPipelines()
	list.setListTotalCount(countPipelines, ctx)

	cursor, err := fromCollection.Aggregate(ctx, dbPipelines)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cursor.Close(ctx)
	list.prepareResult(cursor, ctx, dataSliceType)
}

func (list *List) generateFilterStage(filters map[string]string, columnsMap map[string]Columned) (stage bson.D, remainingFilters map[string]string) {
	filterStage := bson.D{}
	for filterCol, filterValue := range filters {
		if columnInfo, ok := columnsMap[filterCol]; ok {
			filterStage = append(filterStage, bson.E{Key: columnInfo.(*ColumnInfo).Column, Value: columnInfo.getFilterVal(filterValue)})
			delete(filters, filterCol)
		}
	}
	if len(filterStage) > 0 {
		stage = bson.D{{"$match", filterStage}}
	}
	return stage, filters
}

func (list *List) mergePaginationPipelines() (dataPipeline mongo.Pipeline, countPipeline mongo.Pipeline) {
	allColumnsMap := list.fromCollectionColumnsMap

	preparedPipelines := mongo.Pipeline{}
	remainingFilters := make(map[string]string)
	for k, v := range list.Filters {
		remainingFilters[k] = v
	}
	/// add main collection filters
	filterStage, remainingFilters := list.generateFilterStage(remainingFilters, list.fromCollectionColumnsMap)
	if len(filterStage) > 0 {
		preparedPipelines = append(preparedPipelines, filterStage)
	}

	/// add pipelines before sort and pagination stages that maybe change result count
	stages, pipelineIndex, remainingFilters := list.addChangeableResultCountStagesAndFilters(remainingFilters)
	preparedPipelines = append(preparedPipelines, stages...)
	/// add pipelines until remaining filters make empty because pagination stages must add after these pipelines
	for len(remainingFilters) != 0 && pipelineIndex < len(list.dbPipelines) {
		currentStage := list.dbPipelines[pipelineIndex]
		preparedPipelines = append(preparedPipelines, currentStage.Pipeline)
		pipelineIndex++
		if len(currentStage.ResultColumns) > 0 {
			for k, v := range currentStage.ResultColumns {
				allColumnsMap[k] = v
			}
			filterStage, remainingFilters = list.generateFilterStage(remainingFilters, currentStage.ResultColumns)
			if len(filterStage) > 0 {
				preparedPipelines = append(preparedPipelines, filterStage)
			}
		}
	}

	countPipeline = append(preparedPipelines, list.getCountStage())
	// add pagination stages after others stages that can change result count
	preparedPipelines = list.addSortAndPageLimitStages(preparedPipelines, allColumnsMap)

	/// add remaining pipelines
	for i := pipelineIndex; i < len(list.dbPipelines); i++ {
		pipeline := list.dbPipelines[i]
		preparedPipelines = append(preparedPipelines, pipeline.Pipeline)
		if pipeline.ResultColumns != nil {
			preparedPipelines = list.addSortAndPageLimitStages(preparedPipelines, pipeline.ResultColumns)
		}
	}

	/// if client send incorrect sort field pagination stage must add pipelines
	if !list.sortStageAdded {
		log.Printf("filter %s is wrong for sort stage", list.Sort)
		preparedPipelines = append(preparedPipelines, list.getPageLimitStages()...)
	}
	return preparedPipelines, countPipeline
}

func pipelinesHasChangeableResultCount(pipelines []pipeline) bool {
	for _, p := range pipelines {
		if p.ChangeableResultCount {
			return true
		}
	}
	return false
}

func (list *List) addSortAndPageLimitStages(preparedPipelines mongo.Pipeline, columnsMap map[string]Columned) mongo.Pipeline {
	if !list.sortStageAdded {
		if list.Sort == "" {
			preparedPipelines = append(preparedPipelines, list.getPageLimitStages()...)
			list.sortStageAdded = true
			return preparedPipelines
		}
		if sortStage := list.getSortStage(columnsMap); sortStage != nil {
			list.sortStageAdded = true
			preparedPipelines = append(preparedPipelines, sortStage)
			preparedPipelines = append(preparedPipelines, list.getPageLimitStages()...)
		}
	}
	return preparedPipelines
}

func (list *List) getCountStage() bson.D {
	return bson.D{{"$count", "count"}}
}

func (list *List) setListTotalCount(countPipeline mongo.Pipeline, context context.Context) {
	countCursor, err := list.fromCollection.Aggregate(context, countPipeline)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer countCursor.Close(context)
	countCursor.Next(context)
	countData := bson.M{}
	countCursor.Decode(&countData)
	if count, ok := countData["count"]; ok {
		list.TotalItemsCount = int(count.(int32))
	}
}

func (list *List) getPageLimitStages() []bson.D {
	if list.PageNo == 1 {
		return []bson.D{{{"$limit", list.PageSize}}}
	}
	return []bson.D{{{"$skip", (list.PageNo - 1) * list.PageSize}}, {{"$limit", list.PageSize}}}

}

func (list *List) getSortStage(columnsMap map[string]Columned) (sortStage bson.D) {
	if list.Sort == "" {
		return nil
	}
	//sorts := strings.Split(pagination.Sort, ",")
	sortItem := list.Sort
	//for _, sortItem := range sorts {
	field, direction := getSortField(sortItem, columnsMap)
	if field == "" {
		return nil
	}
	sortStage = bson.D{{"$sort", bson.D{{field, direction}}}}
	//}
	return sortStage
}
func getSortField(sortItem string, columnsMap map[string]Columned) (sortField string, direction int) {
	direction = 1
	sortField = sortItem
	if sortItem[:1] == "-" {
		direction = -1
		sortField = sortItem[1:]
	}
	if columnInfo, ok := columnsMap[sortField]; ok {
		return columnInfo.(*ColumnInfo).Column, direction
	}
	return "", 0
}

func (list *List) prepareResult(cursor *mongo.Cursor, context context.Context, dataSliceType interface{}) {
	list.prepareJsonResult(cursor, context, dataSliceType)
}

func (list *List) prepareJsonResult(cursor *mongo.Cursor, context context.Context, dataSliceType interface{}) {
	if err := cursor.All(context, dataSliceType); err != nil {
		fmt.Println(err.Error())
		return
	}
	list.data = dataSliceType
}

// addChangeableResultCountStagesAndFilters if stage has changeable ResultCount flag then it add to prepared pipeline
// and after each state should add filter stage if filter parameter matched with ResultColumn
// @return pipelines, pipelineIndex, remainingFilters
func (list *List) addChangeableResultCountStagesAndFilters(remainingFilters map[string]string) ([]bson.D, int, map[string]string) {
	pipelineIndex := 0
	var pipelines []bson.D
	filterStage := bson.D{}
	/// add pipelines before sort and pagination stages that maybe change result count
	for pipelineIndex < len(list.dbPipelines) && pipelinesHasChangeableResultCount(list.dbPipelines[pipelineIndex:]) {
		pipeline := list.dbPipelines[pipelineIndex]
		pipelines = append(pipelines, pipeline.Pipeline)
		if len(remainingFilters) != 0 {
			filterStage, remainingFilters = list.generateFilterStage(remainingFilters, pipeline.ResultColumns)
			if filterStage != nil {
				pipelines = append(pipelines, filterStage)
			}
		}
		pipelineIndex++
	}
	return pipelines, pipelineIndex, remainingFilters
}
