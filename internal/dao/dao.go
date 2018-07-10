package dao

import (
	"bytes"
	"io"
	"text/template"
)

const (
	daoCode = `
	func (m *{{.StructName}}) TableName() string {
	return m.GetTableKey()
}

func (m *{{.StructName}}) GetDbKey() string {
	return "{{.DbName}}"
}

func (m *{{.StructName}}) GetTableKey() string {
	return "{{ .TableName }}"
}
	
	/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 判断数据是否存在（多个自定义复杂条件查询）
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (m *{{.StructName}}) IsExists(
	query interface{},
	args ...interface{}) (bool, error) {

	if count, err := m.GetCount(query, args...); err != nil {
		if err == gorm.RecordNotFound {
			return false, nil
		}
		return false, err
	} else {
		if count > 0 {
			return true, nil
		}
	}

	return false, nil
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 获取数据记录数（多个自定义复杂条件查询）
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (m *{{.StructName}}) GetCount(
	query interface{},
	args ...interface{}) (int64, error) {
	m.ModelDbMap(m)
	defer m.DbMap.Close()

	paging := ging.Paging{
		PagingIndex: 1,
		PagingSize:  1,
	}

	if err := m.DbMap.Model(m).Where(query, args...).Count(
		&paging.TotalRecord).Error; err != nil {
		return 0, err
	}

	return paging.TotalRecord, nil
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 获取单条数据（单个简单条件查询）
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (m *{{.StructName}}) Select(fieldName string, fieldValue interface{}) error {
	m.ModelDbMap(m)
	defer m.DbMap.Close()

	query := map[string]interface{}{}
	query[fieldName] = fieldValue

	if err := m.DbMap.Find(m, query).Error; err != nil {
		if err == gorm.RecordNotFound {
			return common.NotFoundError
		}
		return err
	}

	return nil
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 获取单条数据（多个自定义复杂条件查询）
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (m *{{.StructName}}) SelectQuery(query interface{}, args ...interface{}) error {
	m.ModelDbMap(m)
	defer m.DbMap.Close()

	if err := m.DbMap.Where(
		query, args...).Find(m).Error; err != nil {
		if err == gorm.RecordNotFound {
			return common.NotFoundError
		}
		return err
	}

	return nil
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 获取单条数据（多个自定义复杂条件查询）
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (m *{{.StructName}}) SelectOrderQuery(
	query interface{},
	sortorder string,
	args ...interface{}) error {
	m.ModelDbMap(m)
	defer m.DbMap.Close()

	if err := m.DbMap.Where(
		query, args...).Order(sortorder).Limit(1).Find(m).Error; err != nil {
		if err == gorm.RecordNotFound {
			return common.NotFoundError
		}
		return err
	}

	return nil
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 获取单条数据（主键标识简单条件查询）
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (m *{{.StructName}}) SelectById(id interface{}) error {
	if err := model.GetFromCache(id.(string), m); err == nil {
		if len(m.Id) > 0 {
			return nil
		}
	}

	m.ModelDbMap(m)
	defer m.DbMap.Close()

	params := map[string]interface{}{
		"id": id,
	}

	if err := m.DbMap.Find(m, params).Error; err != nil {
		if err == gorm.RecordNotFound {
			return common.NotFoundError
		}
		return err
	}

	model.AddToCache(m.Id, m)

	return nil
}

/* ================================================================================
 * query: []interface{} || map[string]interface{} || string
 * args: if string: interface{}...
 * ================================================================================ */
func (m *{{.StructName}}) SelectAll(
	paging *ging.Paging,
	query interface{},
	args ...interface{},
) ([]*{{.StructName}}, error) {
	m.ModelDbMap(m)
	defer m.DbMap.Close()

	var mList []*{{.StructName}} = make([]*{{.StructName}}, 0)
	var err error = nil

	if paging != nil {
		isTotalRecord := true
		if paging.IsTotalOnce {
			if paging.PagingIndex > 1 {
				isTotalRecord = false
			}
		}

		if isTotalRecord && paging.PagingSize > 0 {
			if len(paging.Group) == 0 {
				err = m.DbMap.Model(m).
					Where(query, args...).
					Count(&paging.TotalRecord).
					Order(paging.Sortorder).
					Offset(paging.Offset()).
					Limit(paging.PagingSize).
					Find(&mList).Error
			} else {
				err = m.DbMap.Model(m).
					Where(query, args...).
					Group(paging.Group).
					Count(&paging.TotalRecord).
					Order(paging.Sortorder).
					Offset(paging.Offset()).
					Limit(paging.PagingSize).
					Find(&mList).Error
			}

			paging.SetTotalRecord(paging.TotalRecord)
		} else {
			if len(paging.Group) == 0 {
				err = m.DbMap.Model(m).
					Where(query, args...).
					Order(paging.Sortorder).
					Find(&mList).Error
			} else {
				err = m.DbMap.Model(m).
					Where(query, args...).
					Group(paging.Group).
					Order(paging.Sortorder).
					Find(&mList).Error
			}
		}
	} else {
		err = m.DbMap.Where(query, args...).Find(&mList).Error
	}

	if err != nil {
		if err == gorm.RecordNotFound {
			err = common.NotFoundError
		}
	}

	return mList, err
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 插入数据
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (m *{{.StructName}}) Insert() error {
	if m.DbMap == nil {
		m.ModelDbMap(m)
		defer m.DbMap.Close()
	}

	if err := m.DbMap.Create(m).Error; err != nil {
		return err
	}

	return nil
}

/* ================================================================================
 * data type:
 * Model{"fieldName":"value"...}
 * map[string]interface{}
 * key1,value1,key2,value2
 * ================================================================================ */
func (m *{{.StructName}}) Update(data ...interface{}) (int64, error) {
	if len(m.Id) == 0 || len(data) == 0 {
		return 0, common.ArgumentError
	}

	if m.DbMap == nil {
		m.ModelDbMap(m)
		defer m.DbMap.Close()
	}

	dbContext := m.DbMap.Model(m).UpdateColumns(data)
	rowsAffected, err := dbContext.RowsAffected, dbContext.Error

	if err == nil {
		model.RemoveFromCache(m.Id)
	}

	return rowsAffected, err
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 删除数据
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (m *{{.StructName}}) Delete() (int64, error) {
	if len(m.Id) == 0 {
		return 0, common.ArgumentError
	}

	if m.DbMap == nil {
		m.ModelDbMap(m)
		defer m.DbMap.Close()
	}

	dbContext := m.DbMap.Delete(m)
	rowsAffected, err := dbContext.RowsAffected, dbContext.Error

	if err == nil {
		model.RemoveFromCache(m.Id)
	}

	return rowsAffected, err
}
	`
)

type fillData struct {
	StructName string
	TableName  string
	DbName string
}

// GenerateDao generates Dao code
func GenerateDao(dbName,tableName, structName string) (io.Reader, error) {
	var buff bytes.Buffer
	err := template.Must(template.New("dao").Parse(daoCode)).Execute(&buff, fillData{
		DbName:dbName,
		StructName: structName+"Model",
		TableName:  tableName,
	})
	if nil != err {
		return nil, err
	}
	return &buff, nil
}
