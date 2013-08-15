// red project red.go
package red

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
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

//解析STRUCT字段和值到一个MAP中
func (orm *Orm) scanStructIntoMap(o interface{}) error {
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

//把MAP中的值映射到STRUCT相应字段上
func (orm *Orm) scanMapIntoStruct(o interface{}, omap map[string][]byte) error {
	dataStruct := reflect.Indirect(reflect.ValueOf(o))
	for key, _ := range orm.StructMap {
		for k, data := range omap {
			str := fmt.Sprintf("_%v", strings.ToLower(key))
			if k == str {
				structField := dataStruct.FieldByName(key)
				if !structField.CanSet() {
					continue
				}

				var v interface{}

				switch structField.Type().Kind() {
				case reflect.Slice:
					v = data
				case reflect.String:
					v = string(data)
				case reflect.Bool:
					v = string(data) == "1"
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
					x, err := strconv.Atoi(string(data))
					if err != nil {
						return errors.New("arg " + key + " as int: " + err.Error())
					}
					v = x
				case reflect.Int64:
					x, err := strconv.ParseInt(string(data), 10, 64)
					if err != nil {
						return errors.New("arg " + key + " as int: " + err.Error())
					}
					v = x
				case reflect.Float32, reflect.Float64:
					x, err := strconv.ParseFloat(string(data), 64)
					if err != nil {
						return errors.New("arg " + key + " as float64: " + err.Error())
					}
					v = x
				case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					x, err := strconv.ParseUint(string(data), 10, 64)
					if err != nil {
						return errors.New("arg " + key + " as int: " + err.Error())
					}
					v = x
				case reflect.Struct:
					x, _ := time.Parse("2006-01-02 15:04:05.000 -0700", string(data))
					v = x
				default:
					return errors.New("unsupported type in Scan: " + reflect.TypeOf(v).String())
				}

				structField.Set(reflect.ValueOf(v))
				fmt.Printf("key:%v\n", structField)

			}
		}
	}

	return nil
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
	err := orm.scanStructIntoMap(o)
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

	_, err = orm.Exec()
	if err != nil {
		return err
	}
	return nil
}

//更新数据
func (orm *Orm) Update(o interface{}) error {
	err := orm.scanStructIntoMap(o)
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
	err := orm.scanStructIntoMap(o)
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
	err := orm.scanStructIntoMap(o)
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

	var whereStrName string
	var whereValue interface{}
	for k, v := range whereMap {
		whereStrName = k
		whereValue = v
	}

	orm.SqlStr = fmt.Sprintf("SELECT %v FROM %v WHERE %v=$1", selectStr, orm.TableName, fmt.Sprintf("_%v", strings.ToLower(whereStrName)))

	results, err := orm.query(orm.SqlStr, whereValue)
	if err != nil {
		return err
	}

	err = orm.scanMapIntoStruct(o, results[0])
	if err != nil {
		return err
	}

	return nil
}

func (orm *Orm) query(str string, args ...interface{}) (resultsSlice []map[string][]byte, err error) {
	rows, err := orm.Db.Query(str, args...)
	if err != nil {
		return nil, err
	}

	fields, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		results := make(map[string][]byte)
		var fieldValues []interface{}
		for i := 0; i < len(fields); i++ {
			var fieldValue interface{}
			fieldValues = append(fieldValues, &fieldValue)
		}

		err = rows.Scan(fieldValues...)
		if err != nil {
			return nil, err
		}

		for k, v := range fields {
			fValue := reflect.Indirect(reflect.ValueOf(fieldValues[k])).Interface()
			rowType := reflect.TypeOf(fValue)
			rowValue := reflect.ValueOf(fValue)
			var str string
			switch rowType.Kind() {
			case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				str = strconv.FormatInt(int64(rowValue.Int()), 10)
				results[v] = []byte(str)
			case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				str = strconv.FormatUint(uint64(rowValue.Uint()), 10)
				results[v] = []byte(str)
			case reflect.Float32, reflect.Float64:
				str = strconv.FormatFloat(float64(rowValue.Float()), 'f', -1, 64)
				results[v] = []byte(str)
			case reflect.Slice:
				if rowType.Elem().Kind() == reflect.Uint8 {
					results[v] = fValue.([]byte)
					break
				}
			case reflect.String:
				str = rowValue.String()
				results[v] = []byte(str)
			case reflect.Struct:
				str = fValue.(time.Time).Format("2006-01-02 15:04:05.000 -0700")
				results[v] = []byte(str)
			}
		}
		resultsSlice = append(resultsSlice, results)

	}
	return resultsSlice, nil
}

//取得多条记录
func (orm *Orm) FindList(o interface{}) error {
	err := orm.scanStructIntoMap(o)
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

	var str string

	if len(whereMap) > 0 {
		var whereNames []string
		var whereValues []interface{}
		for k, v := range whereMap {
			whereNames = append(whereNames, k)
			whereValues = append(whereValues, v)
		}
		var whereStrs []string
		for i := 0; i < len(whereNames); i++ {
			whereStrs = append(whereStrs, fmt.Sprintf("%v=$%v", fmt.Sprintf("_%v", strings.ToLower(whereNames[i])), i+1))
		}
		whereStr := strings.Join(whereStrs, " AND ")
		str = fmt.Sprintf("SELECT %v FROM %v WHERE %v", selectStr, orm.TableName, whereStr)
	} else {
		str = fmt.Sprintf("SELECT %v FROM %v", selectStr, orm.TableName)
	}
	fmt.Printf("str:%v\n", str)

	//orm.SqlStr = fmt.Sprintf("SELECT %v FROM %v WHERE %v=$1", selectStr, orm.TableName, fmt.Sprintf("_%v", strings.ToLower(whereStrName)))
	/*
		results, err := orm.query(orm.SqlStr, whereValue)
		if err != nil {
			return err
		}

		err = orm.scanMapIntoStruct(o, results[0])
		if err != nil {
			return err
		}
	*/
	return nil
}
