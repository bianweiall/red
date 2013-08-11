// red project red.go
package red

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type Orm struct {
	Db           *sql.DB
	TableName    string
	TableItemStr string
	ValueItemStr string
	PKName       string
	DateTimeKey  []string
	SqlStr       string
	ParamValue   []interface{}
	StructMap    map[string]interface{}
	FilterStr    []string
}

//设置数据库表名
func (orm *Orm) SetTableName(tabeName string) *Orm {
	orm.TableName = tabeName
	return orm
}

func (orm *Orm) Filter(strs ...string) *Orm {
	orm.FilterStr = strs
	return orm
}

func getStructMap(o interface{}) (map[string]interface{}, []string, error) {
	t := reflect.TypeOf(o).Elem()
	if t.Kind() != reflect.Struct {
		return nil, nil, errors.New("SetSqlStr(o interface)需要传入一个struct类型的指针，例如SetSqlStr(&user)")
	}
	v := reflect.ValueOf(o).Elem()
	var args = make(map[string]interface{})
	var dateTimes []string
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Tag == "dt" {
			dateTimes = append(dateTimes, t.Field(i).Name)
		}
		args[t.Field(i).Name] = v.Field(i).Interface()
	}
	return args, dateTimes, nil
}

func (orm *Orm) Exec() (sql.Result, error) {
	fmt.Printf("orm.SqlStr:%v\n", orm.SqlStr)
	//stmt, err := orm.Db.Prepare(orm.SqlStr)
	if err != nil {
		return nil, errors.New("Prepare错误")
	}
	defer stmt.Close()
	//fmt.Printf("orm.ParamValue:%v\n", orm.ParamValue)
	res, err := stmt.Exec(orm.ParamValue...)
	if err != nil {
		return nil, errors.New("Exec错误")
	}
	return res, nil
}

func isInMap(args map[string]interface{}, strs []string) error {
	var names []string
	for n, _ := range args {
		names = append(names, n)
	}
	var filterLen = len(strs)

	for i := 0; i < filterLen; i++ {
		if strings.Contains(strings.Join(names, ","), strs[i]) != true {
			return errors.New("Filter输入的字符串不是struct中的字段!")
		}
	}
	return nil
}

func (orm *Orm) Create(o interface{}) error {
	args, dt, err := getStructMap(o)
	if err != nil {
		return err
	}

	err = isInMap(args, orm.FilterStr)
	if err != nil {
		return err
	}

	for j := 0; j < len(orm.FilterStr); j++ {
		delete(args, orm.FilterStr[j])
	}

	for j := 0; j < len(dt); j++ {
		delete(args, dt[j])
	}

	var names []string
	var namevlues []string
	var values []interface{}

	for n, v := range args {
		names = append(names, n)
		values = append(values, v)
	}

	for j := 1; j <= len(names); j++ {
		namevlues = append(namevlues, fmt.Sprintf("$%v", j))
	}

	for j := 0; j < len(dt); j++ {
		names = append(names, dt[j])
		namevlues = append(namevlues, "current_timestamp")
	}

	orm.TableItemStr = "_" + strings.ToLower(strings.Join(names, ",_"))
	orm.ValueItemStr = strings.Join(namevlues, ",")
	orm.ParamValue = values

	orm.SqlStr = fmt.Sprintf("INSERT INTO %v(%v) VALUES(%v)", orm.TableName, orm.TableItemStr, orm.ValueItemStr)

	_, err = orm.Exec()
	if err != nil {
		return err
	}
	return nil
}
