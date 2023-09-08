package models

import "gorm.io/gorm"

type DiggerJob struct {
	gorm.Model
	DiggerJobId       string  `gorm:"size:50,index:idx_digger_job_id"`
	ParentDiggerJobId *string `gorm:"size:50,index:idx_parent_digger_job_id"`
	Status            DiggerJobStatus
}