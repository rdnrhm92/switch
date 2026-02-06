package dto

import "time"

type PageLimit struct {
	Page     uint `json:"page" `
	PageSize uint `json:"pageSize"`
	Offset   uint
	Limit    uint
}

func (p *PageLimit) ComputeLimit() {
	if p.Page == 0 {
		p.Page = 1
	}
	if p.PageSize == 0 {
		p.PageSize = 20
	}
	p.Offset = (p.Page - 1) * p.PageSize
	p.Limit = p.PageSize
}

type PageTime struct {
	StartTime *time.Time `json:"startTime"`
	EndTime   *time.Time `json:"endTime"`
}
