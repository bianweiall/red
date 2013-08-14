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
	//需要提交到数据库字段的值
	ParamValue []interface{}
	//数据库主键，`pk:auto`为自动递增主键，`pk`为不能自动递增主键
	PKName string
	//自动主键
	AutoPKName string
	//tag为`dt`的字段名
	DateTimeKey []string
	//解析struct为一个map，去除了包含"pk"或者是"pk:auto"或者是"dt"的字段和值
	StructMap map[string]interface{}
	//过滤字符串
	FilterStrs []string

	WhereMap map[string]interface{}
}

//设置数据库表名
func (orm *Orm) SetTableName(tabeName string) *Orm {
	orm.TableName = tabeName
	return orm
}

//设置过滤条件
func (orm *Orm) Filter(args ...string) *Orm {
	orm.FilterStrs = args
	return orm
}

//设置Where条件，例如：Where("Id,name",1,"学友")
func (orm *Orm) Where(str string, args ...interface{}) *Orm {
	strs := strings.Split(str, ",")
	var myWhere = make(map[string]interface{})
	for i := 0; i < len(strs); i++ {
		myWhere[strings.TrimSpace(strs[i])] = args[i]
	}
	orm.WhereMap = myWhere
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
		args[t.Field(i).Name] = v.Field(i).Interface()

		if t.Field(i).Tag == "pk:auto" {
			pkName = t.Field(i).Name
			orm.AutoPKName = pkName
			j++

		} else if t.Field(i).Tag == "pk" {
			pkName = t.Field(i).Name
			j++
		} else if t.Field(i).Tag == "dt" {
			dateTimes = append(dateTimes, t.Field(i).Name)
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
	pk := orm.PKName

	delete(args, orm.PKName)

	for j := 0; j < len(dt); j++ {
		delete(args, dt[j])
	}

	fmt.Printf("args:%v\n", args)
	fmt.Printf("dt:%v\n", dt)
	fmt.Printf("pk:%v\n", pk)

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
	tableItemStr := fmt.Sprintf("_%v", strings.ToLower(strings.Join(names, ",_")))
	valueItemStr := strings.Join(namevlues, ",")

	orm.ParamValue = values
	orm.SqlStr = fmt.Sprintf("INSERT INTO %v(%v) VALUES(%v)", orm.TableName, tableItemStr, valueItemStr)
	fmt.Printf("values:%v\n", values)
	fmt.Printf("orm.SqlStr:%v\n", orm.SqlStr)

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

	delete(args, orm.AutoPKName)

	for j := 0; j < len(dt); j++ {
		delete(args, dt[j])
	}

	var names []string
	var namevlues []string
	var values []interface{}

	var snum = 1
	for k, v := range args {
		names = append(names, k)
		namevlues = append(namevlues, fmt.Sprintf("$%v", snum))
		values = append(values, v)
		snum++
	}

	whereMap := orm.WhereMap

	if len(whereMap) != 1 {
		return errors.New("WHERE条件参数只能为一个")
	}

	var whereStrName string
	var whereValue interface{}
	for k, v := range whereMap {
		whereStrName = k
		whereValue = v
	}
	values = append(values, whereValue)

	for j := 0; j < len(dt); j++ {
		if strings.ToLower(dt[j]) != strings.ToLower(whereStrName) {
			names = append(names, dt[j])
			namevlues = append(namevlues, "current_timestamp")
		}
	}

	var setStrs []string
	for i := 0; i < len(names); i++ {
		setStrs = append(setStrs, fmt.Sprintf("%v=%v", fmt.Sprintf("_%v", strings.ToLower(names[i])), namevlues[i]))
	}

	setStr := strings.Join(setStrs, ",")
	whereStr := fmt.Sprintf("%v=%v", fmt.Sprintf("_%v", strings.ToLower(whereStrName)), fmt.Sprintf("$%v", snum))

	orm.ParamValue = values
	orm.SqlStr = fmt.Sprintf("UPDATE %v SET %v WHERE %v", orm.TableName, setStr, whereStr)

	_, err = orm.Exec()
	if err != nil {
		return err
	}
	return nil
}

//删除数据
func (orm *Orm) Delete(o interface{}) error {
	err := orm.getStructMap(o)
	if err != nil {
		return err
	}

	whereMap := orm.WhereMap

	if len(whereMap) != 1 {
		return errors.New("WHERE条件参数只能为一个")
	}
	var values []interface{}
	var whereStrName string
	var whereValue interface{}
	for k, v := range whereMap {
		whereStrName = k
		whereValue = v
	}
	values = append(values, whereValue)
	orm.ParamValue = values
	orm.SqlStr = fmt.Sprintf("DELETE FROM %v WHERE %v=$1", orm.TableName, fmt.Sprintf("_%v", strings.ToLower(whereStrName)))

	_, err = orm.Exec()
	if err != nil {
		return err
	}
	return nil
}

//取得一条记录
func (orm *Orm) FindOne(o interface{}) error {
	err := orm.getStructMap(o)
	if err != nil {
		return err
	}

	args := orm.StructMap
	fs := orm.FilterStrs

	for i := 0; i < len(fs); i++ {
		delete(args, fs[i])
	}

	var selectStrs []string
	for k, _ := range args {
		selectStrs = append(selectStrs, fmt.Sprintf("_%v", strings.ToLower(k)))
	}

	selectStr := strings.Join(selectStrs, ",")

	whereMap := orm.WhereMap

	if len(whereMap) != 1 {
		return errors.New("WHERE条件参数只能为一个")
	}
	var values []interface{}
	var whereStrName string
	var whereValue interface{}
	for k, v := range whereMap {
		whereStrName = k
		whereValue = v
	}
	values = append(values, whereValue)
	orm.ParamValue = values
	orm.SqlStr = fmt.Sprintf("SELECT %v FROM %v WHERE %v=$1", selectStr, orm.TableName, fmt.Sprintf("_%v", strings.ToLower(whereStrName)))

	rows, err := orm.Db.Query(orm.SqlStr)
	if err != nil {
		return err
	}

	for rows.Next() {
		result := make(map[string][]byte)
		//var c Category
		err = rows.Scan(&id, &title, &author, &content, &createdtime, &cid)

		if err != nil {
			return err
		}
		//fmt.Printf("cid:%v\n", cid)
		b.Id = id
		b.Title = title
		b.Author = author
		b.Content = content
		b.CreatedTime = createdtime
		b.CreatedTime = formatTime(b.CreatedTime)
		b.Category.Id = cid

	}
	return nil
}
