package dataset

/*
Request
	AdvId
	UserId
	Offset
	Limit
	SortBy
		CreateTime
		TotalEvent
		Adspending
Response
	[]Items
	TotalCount
	Offset
	Limit
	SortBy
*/
func ListDataSets() {
	1. offset > 0 limit < 100
	2. 先查所有的下游获取这个 adv 所有 dataset
	3. 再根据 sortby
		查 统计 获取 event
		查 表 获取 cost
	4. 按照 sortby 排序后用 stream.drop.take 获取分区结果
	// 查出所有的数据
	// 按照sortby排序
	// 取offset+limit
}

/*
Request
	AdvId
	UserId
	[]Source
Response
	[]QuotoInfo
		Source
		Limit(from TCC)
		Used
		Free
*/
func CheckQuotoInfo() {

}
