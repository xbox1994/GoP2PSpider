package types

type QueryParam struct {
	Q     string
	Start int
}

type QueryResult struct {
	Hits      int64
	Start     int
	Query     string
	PrevStart int
	NextStart int
	Items     []interface{}
}
