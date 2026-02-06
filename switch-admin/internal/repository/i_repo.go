package repository

import (
	"reflect"

	"gorm.io/gorm"
)

type IRepository interface {
	GetDB() *gorm.DB
}

type BaseRepository struct {
	DB *gorm.DB
}

func (b *BaseRepository) GetDB() *gorm.DB {
	return b.DB
}

// WithTx 事务生成
func WithTx[T IRepository](repo T, tx *gorm.DB) T {
	repoType := reflect.TypeOf(repo).Elem()
	newRepoValue := reflect.New(repoType)

	baseRepoField := newRepoValue.Elem().FieldByName("BaseRepository")
	if !baseRepoField.IsValid() || !baseRepoField.CanSet() {
		panic("repository struct must embed BaseRepository")
	}

	dbField := baseRepoField.FieldByName("DB")
	if !dbField.IsValid() || !dbField.CanSet() {
		panic("BaseRepository must have an exported 'DB' field of type *gorm.DB")
	}

	dbField.Set(reflect.ValueOf(tx))

	return newRepoValue.Interface().(T)
}
