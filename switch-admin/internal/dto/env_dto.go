package dto

// EnvListReq 列表的查询
type EnvListReq struct {
	*PageLimit
	*PageTime
	Name         string `json:"name"`
	Tag          string `json:"tag"`
	Description  string `json:"description"`
	CreatedBy    string `json:"createdBy"`
	NamespaceTag string `json:"namespace_tag"`
}

func (n *EnvListReq) Initial() {
	if n.PageLimit != nil {
		n.PageLimit.ComputeLimit()
	}
}
