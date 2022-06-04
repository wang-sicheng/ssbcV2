package gen

type exists struct{}
type set struct {
	m map[interface{}]exists
}

func NewSet(items ...interface{}) *set {
	s := &set{}
	s.m = make(map[interface{}]exists)
	s.Add(items...)
	return s
}
func (s *set) Add(items ...interface{}) {
	for _, item := range items {
		s.m[item] = exists{}
	}
}
func (s *set) Remove(item interface{}) {
	delete(s.m, item)
}
func (s *set) Contains(item interface{}) bool {
	_, ok := s.m[item]
	return ok
}

// DataSlice 获取所有元素的列表
func (s *set) DataSlice() []interface{} {
	var retList []interface{}
	for ele, _ := range s.m {
		retList = append(retList, ele)
	}
	return retList
}
