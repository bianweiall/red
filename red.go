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
	//SQL中参数的值
	ParamValues []interface{}
	//数据库主键，`pk:auto`为自动递增主键，`pk`为不能自动递增主键
	PKName string
	//自动主键
	AutoPKName string
	//tag为`dt`的字段名
	DateTimeNames []string
	//解析struct为一个map，去除了包含"pk"或者是"pk:auto"或者是"dt"的字段和值
	StructMap map[string]interface{}
	//过滤字符串
	FilterStrs []string

	OrderByStr string

	LimitStr int

	OffsetStr int

	WhereStr      string
	WhereStrValue interface{}

	WhereOrStrs       []string
	WhereOrStrsValues []interface{}

	WhereAndStrs       []string
	WhereAndStrsValues []interface{}
}

func NewOrm() *Orm {
	orm := Orm{
		SqlStr:        "",
		TableName:     "",
		ParamValues:   make([]interface{}, 0),
		PKName:        "",
		AutoPKName:    "",
		DateTimeNames: make([]string, 0),
		StructMap:     make(map[string]interface{}),
		FilterStrs:    make([]string, 0),
		OrderByStr:    "",
		LimitStr:      0,
		OffsetStr:     0,
		WhereStr:      "",
		//WhereStrValue:      nil,
		WhereOrStrs:        make([]string, 0),
		WhereOrStrsValues:  make([]interface{}, 0),
		WhereAndStrs:       make([]string, 0),
		WhereAndStrsValues: make([]interface{}, 0)}
	return &orm
}

//设置数据库表名
func (orm *Orm) SetTableName(tabeName string) *Orm {
	orm.TableName = tabeName
	return orm
}

//设置过滤条件
func (orm *Orm) Filter(filter string) *Orm {
	var strs []string
	if strings.Contains(filter, ",") == true {
		strs = strings.Split(filter, ",")
		orm.FilterStrs = strs
	} else {
		orm.FilterStrs = append(strs, filter)
	}
	return orm
}

//设置Where条件
func (orm *Orm) Where(str string, strValue interface{}) *Orm {
	if str != "" && strValue != nil {
		orm.WhereStr = str
		orm.WhereStrValue = strValue
	}
	return orm
}

//设置WhereOr条件
func (orm *Orm) WhereOr(str string, strValues ...interface{}) *Orm {
	var names []string
	var values []interface{}
	if strings.Contains(str, ",") == true {
		strs := strings.Split(str, ",")
		if len(strs) == len(strValues) {
			for i := 0; i < len(strs); i++ {
				names = append(names, strs[i])
				values = append(values, strValues[i])
			}
		}
	} else {
		if str != "" && len(strValues) == 1 {
			names = append(names, str)
			values = append(values, strValues[0])
		}
	}
	orm.WhereOrStrs = names
	orm.WhereOrStrsValues = values
	return orm
}

//设置WhereAnd条件
func (orm *Orm) WhereAnd(str string, strValues ...interface{}) *Orm {
	var names []string
	var values []interface{}
	if strings.Contains(str, ",") == true {
		strs := strings.Split(str, ",")
		if len(strs) == len(strValues) {
			for i := 0; i < len(strs); i++ {
				names = append(names, strs[i])
				values = append(values, strValues[i])
			}
		}
	} else {
		if str != "" && len(strValues) == 1 {
			names = append(names, str)
			values = append(values, strValues[0])
		}
	}
	orm.WhereAndStrs = names
	orm.WhereAndStrsValues = values
	return orm
}

//设置ORDER BY
func (orm *Orm) OrderBy(orderByStrs ...string) *Orm {
	var strs []string
	var orderByStr string

	if len(orderByStrs) > 0 {
		for _, v := range orderByStrs {
			if strings.Contains(v, "-") == true {
				strs = append(strs, fmt.Sprintf("%v DESC", fmt.Sprintf("_%v", strings.ToLower(strings.TrimLeft(v, "-")))))
			} else {
				strs = append(strs, fmt.Sprintf("%v ASC", fmt.Sprintf("_%v", strings.ToLower(v))))
			}
		}
	}

	if len(orderByStrs) == 1 {
		orderByStr = fmt.Sprintf("ORDER BY %v", strs[0])
	} else {
		orderByStr = fmt.Sprintf("ORDER BY %v", strings.Join(strs, ", "))
	}

	orm.OrderByStr = orderByStr

	return orm
}

//设置Limit
func (orm *Orm) Limit(limit int) *Orm {
	if limit > 0 {
		orm.LimitStr = limit
	}
	return orm
}

//设置Offset
func (orm *Orm) Offset(offset int) *Orm {
	if offset > 0 {
		orm.OffsetStr = offset
	}
	return orm
}

//持久化到数据库
func (orm *Orm) Exec() (sql.Result, error) {
	stmt, err := orm.Db.Prepare(orm.SqlStr)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	res, err := stmt.Exec(orm.ParamValues...)
	if err != nil {
		return nil, err
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
	dt := orm.DateTimeNames

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

	orm.ParamValues = values
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
	dt := orm.DateTimeNames

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

	whereStrName, whereValue := orm.WhereStr, orm.WhereStrValue
	if whereStrName == "" || whereValue == nil {
		return errors.New("where条件NAME不能为空,VALUE不能为nil")
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

	orm.ParamValues = values
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

	whereStrName, whereValue := orm.WhereStr, orm.WhereStrValue
	if whereStrName == "" || whereValue == nil {
		return errors.New("where条件NAME不能为空,VALUE不能为nil")
	}

	var values []interface{}
	values = append(values, whereValue)
	orm.ParamValues = values
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

	whereStrName, whereValue := orm.WhereStr, orm.WhereStrValue
	if whereStrName == "" || whereValue == nil {
		return errors.New("where条件NAME不能为空,VALUE不能为nil")
	}

	orm.SqlStr = fmt.Sprintf("SELECT %v FROM %v WHERE %v=$1", selectStr, orm.TableName, fmt.Sprintf("_%v", strings.ToLower(whereStrName)))
	var wValue []interface{}
	wValue = append(wValue, whereValue)
	results, err := orm.query(orm.SqlStr, wValue)
	if err != nil {
		return err
	}

	err = orm.ScanMapIntoStruct(o, results[0])
	if err != nil {
		return err
	}

	return nil
}

//取得1条记录或多条记录
func (orm *Orm) Find(slicePtr interface{}) error {
	sliceValue := reflect.Indirect(reflect.ValueOf(slicePtr))
	if sliceValue.Kind() != reflect.Slice {
		return errors.New("需要接收一个指针类型的Slice")
	}

	sliceElementType := sliceValue.Type().Elem()

	st := reflect.New(sliceElementType)
	err := orm.scanStructIntoMap(st.Interface())
	if err != nil {
		return err
	}

	selectSql, values := orm.getSelectSqlAndValues()
	resultsSlice, err := orm.query(selectSql, values)
	if err != nil {
		return err
	}

	for _, results := range resultsSlice {
		newValue := reflect.New(sliceElementType)
		err = orm.ScanMapIntoStruct(newValue.Interface(), results)
		if err != nil {
			return err
		}
		sliceValue.Set(reflect.Append(sliceValue, reflect.Indirect(reflect.ValueOf(newValue.Interface()))))
	}

	return nil
}

func (orm *Orm) query(str string, args []interface{}) ([]map[string][]byte, error) {
	rows, err := orm.Db.Query(str, args...)
	if err != nil {
		return nil, err
	}

	fields, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	var resultsSlice []map[string][]byte
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

func (orm *Orm) getSelectSqlAndValues() (string, []interface{}) {
	var allValues []interface{}
	var whereSqlStr, str1, str2 string
	j := 1

	if orm.WhereStr == "" {
		whereSqlStr = ""
	} else {
		whereOrStrs := orm.WhereOrStrs
		if len(whereOrStrs) < 1 {
			str1 = fmt.Sprintf("WHERE %v=$1", fmt.Sprintf("_%v", strings.ToLower(orm.WhereStr)))
			allValues = append(allValues, orm.WhereStrValue)
		} else {
			allValues = append(allValues, orm.WhereStrValue)
			for i := 0; i < len(whereOrStrs); i++ {
				whereOrStrs[i] = fmt.Sprintf("%v=$%v", fmt.Sprintf("_%v", strings.ToLower(whereOrStrs[i])), j+1)
				allValues = append(allValues, orm.WhereOrStrsValues[i])
				j++
			}
			str1 = fmt.Sprintf("WHERE (%v=$1 OR %v)", fmt.Sprintf("_%v", strings.ToLower(orm.WhereStr)), strings.Join(whereOrStrs, " OR "))
		}
		whereAndStrs := orm.WhereAndStrs
		if len(whereAndStrs) < 1 {
			str2 = ""
		} else {
			for i := 0; i < len(whereAndStrs); i++ {
				whereAndStrs[i] = fmt.Sprintf("%v=$%v", fmt.Sprintf("_%v", strings.ToLower(whereAndStrs[i])), j+1)
				allValues = append(allValues, orm.WhereAndStrsValues[i])
				j++
			}
			str2 = fmt.Sprintf("AND %v", strings.Join(whereAndStrs, " AND "))
		}
		whereSqlStr = fmt.Sprintf("%v %v", str1, str2)
	}

	//LIMIT OFFSET
	var limitStr, offsetStr string
	if orm.LimitStr < 1 {
		limitStr = ""
	} else {
		limitStr = fmt.Sprintf("LIMIT $%v", j+1)
		var l interface{}
		l = orm.LimitStr
		allValues = append(allValues, l)
		j++
	}

	if orm.OffsetStr < 1 {
		offsetStr = ""
	} else {
		offsetStr = fmt.Sprintf("OFFSET $%v", j+1)
		var o interface{}
		o = orm.OffsetStr
		allValues = append(allValues, o)
	}

	//SELECT
	args := orm.StructMap
	fs := orm.FilterStrs
	var selectStr string
	if len(fs) < 1 {
		selectStr = "*"
	} else {
		for i := 0; i < len(fs); i++ {
			delete(args, fs[i])
		}
		var selectStrs []string
		for k, _ := range args {
			selectStrs = append(selectStrs, fmt.Sprintf("_%v", strings.ToLower(k)))
		}
		selectStr = strings.Join(selectStrs, ",")
	}

	selectSql := fmt.Sprintf("SELECT %v FROM %v %v %v %v %v", selectStr, orm.TableName, whereSqlStr, orm.OrderByStr, limitStr, offsetStr)
	orm.SqlStr = selectSql
	orm.ParamValues = allValues

	return selectSql, allValues

}

//解析STRUCT字段和值到一个MAP中
func (orm *Orm) scanStructIntoMap(o interface{}) error {
	if reflect.TypeOf(o).Kind() != reflect.Ptr {
		return errors.New("要求传入一个struct类型的指针")
	}
	structValue := reflect.Indirect(reflect.ValueOf(o))
	if structValue.Kind() != reflect.Struct {
		return errors.New("要求传入一个struct类型的指针")
	}

	t := reflect.TypeOf(o).Elem()
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
			return errors.New("要求struct字段只能设置一个主键")
		}
	}

	if orm.TableName == "" {
		orm.TableName = fmt.Sprintf("_%v", strings.ToLower(t.Name()))
	}

	orm.DateTimeNames = dateTimes
	orm.PKName = pkName
	orm.StructMap = args

	return nil
}

//把MAP中的值映射到STRUCT相应字段上
func (orm *Orm) ScanMapIntoStruct(o interface{}, omap map[string][]byte) error {
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

			}
		}
	}

	return nil
}
