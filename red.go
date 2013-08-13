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
	Db *sql.DB
	//需要执行的SQL语句，例如："INSERT INTO %v(%v) VALUES(%v)"，其中3个%v代表Orm类型的TableName,TableItemStr,ValueItemStr字段
	SqlStr string
	//数据库表名
	TableName string
	//SqlStr中表字段名，例如："_name,_age,_time"
	TableItemStr string
	//SqlStr中表字段名的占位符，例如："$1,$2,current_timestamp"
	ValueItemStr string
	//需要提交到数据库字段的值
	ParamValue []interface{}
	//数据库主键，`pk:auto`为自动递增主键，`pk`为不能自动递增主键
	PKName string
	//tag为`dt`的字段名
	DateTimeKey []string
	//解析struct为一个map
	StructMap map[string]interface{}
	//过滤字符串
	FilterStr []string

	WhereStr map[string]interface{}
}

//设置数据库表名
func (orm *Orm) SetTableName(tabeName string) *Orm {
	orm.TableName = tabeName
	return orm
}

//设置过滤条件
func (orm *Orm) Filter(args ...string) *Orm {
	orm.FilterStr = args
	return orm
}

//设置Where条件，例如：Where("Id=$1,Name=$2",1,"学友")
func (orm *Orm) Where(strs ...string, args ...interface{}) *Orm {
	var myWhere = make(map[string][]interface{})
	myWhere["whereKey"] = strs
	myWhere["whereValue"] = args
	orm.WhereStr = myWhere
	return orm
}

//解析传进来的struct信息
func (orm *Orm) getStructMap(o interface{}) error {
	if reflect.TypeOf(o).Kind() != reflect.Ptr {
		return errors.New("只接受指针类型")
	}
	t := reflect.TypeOf(o).Elem()
	if t.Kind() != reflect.Struct {
		return errors.New("只接受struct类型的指针")
	}
	v := reflect.ValueOf(o).Elem()

	var args = make(map[string]interface{})
	var dateTimes []string
	var pkName string
	var j = 0

	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Tag == "pk:auto" {
			pkName = t.Field(i).Name
			j++

		} else if t.Field(i).Tag == "pk" {
			pkName = t.Field(i).Name
			args[t.Field(i).Name] = v.Field(i).Interface()
			j++
		} else if t.Field(i).Tag == "dt" {
			dateTimes = append(dateTimes, t.Field(i).Name)
		} else {
			args[t.Field(i).Name] = v.Field(i).Interface()
		}

		if j > 1 {
			return errors.New("struct字段只能设置一个主键")
		}
	}

	if orm.TableName == "" {
		orm.TableName = fmt.Sprintf("_%v", strings.ToLower(t.Name()))
	}

	orm.DateTimeKey = dateTimes
	orm.PKName = pkName
	orm.StructMap = args

	return nil
}

//判断strs中的每个字符串是否在map的KEY中都能找到，如果都能找到就返回true
func isInMap(args map[string]interface{}, strs []string) bool {
	var names []string
	for n, _ := range args {
		names = append(names, n)
	}
	var filterLen = len(strs)

	for i := 0; i < filterLen; i++ {
		if strings.Contains(strings.Join(names, ","), strs[i]) != true {
			return false
		}
	}
	return true
}

//持久化到数据库
func (orm *Orm) Exec() (sql.Result, error) {
	fmt.Printf("orm.SqlStr:%v\n", orm.SqlStr)
	stmt, err := orm.Db.Prepare(orm.SqlStr)
	if err != nil {
		return nil, errors.New("Prepare错误")
	}
	defer stmt.Close()
	fmt.Printf("orm.ParamValue:%v\n", orm.ParamValue)
	res, err := stmt.Exec(orm.ParamValue...)
	if err != nil {
		return nil, errors.New("Exec错误")
	}
	return res, nil
}

//保存数据
func (orm *Orm) Create(o interface{}) error {
	err := orm.getStructMap(o)
	if err != nil {
		return err
	}
	args := orm.StructMap
	dt := orm.DateTimeKey

	/*
			bok := isInMap(args, orm.FilterStr)
			if bok != true {
				return errors.New("FilerrStr中的字符串不是struct中的字段")
			}

			for j := 0; j < len(orm.FilterStr); j++ {
				delete(args, orm.FilterStr[j])
			}

		for j := 0; j < len(dt); j++ {
			delete(args, dt[j])
		}
	*/

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

	orm.TableItemStr = fmt.Sprintf("_%v", strings.ToLower(strings.Join(names, ",_")))
	orm.ValueItemStr = strings.Join(namevlues, ",")
	orm.ParamValue = values

	orm.SqlStr = fmt.Sprintf("INSERT INTO %v(%v) VALUES(%v)", orm.TableName, orm.TableItemStr, orm.ValueItemStr)

	_, err = orm.Exec()
	if err != nil {
		return err
	}
	return nil
}

//更新数据
func (orm *Orm) Update(o interface{}) error {
	err := orm.getStructMap(o)
	if err != nil {
		return err
	}
	args := orm.StructMap
	dt := orm.DateTimeKey

	var names []string
	var namevlues []string
	var values []interface{}

	var i = 1
	for k, v := range args {
		names = append(names, k)
		namevlues = append(namevlues, fmt.Sprintf("$%v", i))
		values = append(values, v)
		i++
	}

	for j := 0; j < len(dt); j++ {
		names = append(names, dt[j])
		namevlues = append(namevlues, "current_timestamp")
	}
	//var setStrNames []string
	//var setStrNameValues []string
	for j := 0; j < len(names); j++ {
		names[j] = fmt.Sprintf("_%v", strings.ToLower(names[j]))
		//setStrNames = append(setStrNames, fmt.Sprintf("_%v", strings.ToLower(names[j])))
		//setStrNameValues = append(setStrNameValues, fmt.Sprintf("_%v", strings.ToLower(namevlues[j])))
	}

	strs := orm.WhereStr["whereKey"]
	var whereStrs []string
	var whereValueStrs []string
	for j := 0; j < len(strs); j++ {
		//whereStrs = append(whereStrs, strs[j])

		//whereValueStrs = append(whereValueStrs, fmt.Sprintf("$%v", i+1))
	}

	orm.SqlStr = fmt.Sprintf("UPDATE %v SET %v WHERE %v", orm.TableName, setStr)

	/*
		if strings.Contains(fmt.Sprintf("%v", dt), k) == true {
			namevlues = append(namevlues, "current_timestamp")
		} else {
			namevlues = append(namevlues, fmt.Sprintf("$%v", i))
			i++
		}*/
	fmt.Printf("names:%v\n", names)
	fmt.Printf("namevlues:%v\n", namevlues)
	fmt.Printf("values:%v\n", values)

	//orm.TableItemStr = fmt.Sprintf("_%v", strings.ToLower(strings.Join(names, ",_")))
	//orm.ValueItemStr = strings.Join(namevlues, ",")
	//orm.ParamValue = values

	//orm.SqlStr = fmt.Sprintf("UPDATE %v SET title=$1,author=$2,content=$3,cid=$4,createdtime=current_timestamp WHERE id=$5", orm.TableName, orm.TableItemStr, orm.ValueItemStr)

	//_, err = orm.Exec()
	//if err != nil {
	//	return err
	//}
	return nil
}
