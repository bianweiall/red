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

	OrderByStr string

	LimitStr int

	OffsetStr int
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
	nowMap := make(map[string]interface{})
	nowMap[str] = strValue
	orm.WhereMap = nowMap
	return orm
}

//设置WhereOr条件
func (orm *Orm) WhereOr(str string, strValues ...interface{}) *Orm {
	if strings.Contains(str, ",") == true {
		strs := strings.Split(str, ",")
		if len(strs) == len(strValues) {
			for i := 0; i < len(strs); i++ {
				orm.WhereMap[fmt.Sprintf("||%v", strings.TrimSpace(strs[i]))] = strValues[i]
			}
		}
	} else {
		if len(strValues) == 1 {
			orm.WhereMap[fmt.Sprintf("||%v", strings.TrimSpace(str))] = strValues[0]
		}
	}
	return orm
}

//设置WhereAnd条件
func (orm *Orm) WhereAnd(str string, strValues ...interface{}) *Orm {
	if strings.Contains(str, ",") == true {
		strs := strings.Split(str, ",")
		if len(strs) == len(strValues) {
			for i := 0; i < len(strs); i++ {
				orm.WhereMap[fmt.Sprintf("&%v", strings.TrimSpace(strs[i]))] = strValues[i]
			}
		}
	} else {
		if len(strValues) == 1 {
			orm.WhereMap[fmt.Sprintf("&%v", strings.TrimSpace(str))] = strValues[0]
		}
	}
	return orm
}

//设置ORDER BY
func (orm *Orm) OrderBy(orderByStrs ...string) *Orm {
	var strs []string
	var orderByStr string

	if len(orderByStrs) < 1 {
		orderByStr = ""
	} else {
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
	orm.LimitStr = limit
	return orm
}

//设置Offset
func (orm *Orm) Offset(offset int) *Orm {
	orm.OffsetStr = offset
	return orm
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
		return nil, err
	}
	defer stmt.Close()
	fmt.Printf("orm.ParamValue:%v\n", orm.ParamValue)
	res, err := stmt.Exec(orm.ParamValue...)
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

	whereStrName, whereValue, err := getOneInMap(orm.WhereMap)
	if err != nil {
		return err
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

	whereStrName, whereValue, err := getOneInMap(orm.WhereMap)
	if err != nil {
		return err
	}
	var values []interface{}
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

	whereStrName, whereValue, err := getOneInMap(orm.WhereMap)
	if err != nil {
		return err
	}

	orm.SqlStr = fmt.Sprintf("SELECT %v FROM %v WHERE %v=$1", selectStr, orm.TableName, fmt.Sprintf("_%v", strings.ToLower(whereStrName)))
	var wValue []interface{}
	wValue = append(wValue, whereValue)
	results, err := orm.query(orm.SqlStr, wValue)
	if err != nil {
		return err
	}

	err = orm.scanMapIntoStruct(o, results[0])
	if err != nil {
		return err
	}

	return nil
}

//MAP中只有一个key,value时，获取key,value
func getOneInMap(args map[string]interface{}) (string, interface{}, error) {
	if len(args) != 1 {
		return "", nil, errors.New("MAP中只有一个key,value时有用")
	}
	var whereStrName string
	var whereValue interface{}
	for k, v := range args {
		whereStrName = k
		whereValue = v
	}
	return whereStrName, whereValue, nil
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

func getWhereStr(args map[string]interface{}) (string, []interface{}) {
	var whereStr, whereName string
	var whereNameValue interface{}
	var whereOrs, whereAnds []string
	var whereOrValues, whereAndValues, allValues []interface{}
	j := 2
	if len(args) < 1 {
		return "", make([]interface{}, 0)
	} else {
		for k, v := range args {
			if strings.Contains(k, "||") == true {
				whereOrs = append(whereOrs, fmt.Sprintf("_%v=$%v", strings.ToLower(strings.TrimLeft(k, "||")), j))
				whereOrValues = append(whereOrValues, v)
				j++
			} else if strings.Contains(k, "&") == true {
				whereAnds = append(whereAnds, fmt.Sprintf("_%v", strings.ToLower(strings.TrimLeft(k, "&"))))
				whereAndValues = append(whereAndValues, v)
			} else {
				whereName = fmt.Sprintf("_%v=$1", strings.ToLower(k))
				whereNameValue = v
			}
		}

		for i := 0; i < len(whereAnds); i++ {
			whereAnds[i] = fmt.Sprintf("%v=$%v", whereAnds[i], j)
			j++
		}

		//拼接字符串/所有条件的值
		var whereNameAndOrStr string
		var whereAndStr string
		allValues = append(allValues, whereNameValue)
		if len(whereOrValues) == 1 {
			whereNameAndOrStr = fmt.Sprintf("WHERE (%v OR %v)", whereName, whereOrs[0])
			allValues = append(allValues, whereOrValues[0])
		} else if len(whereOrValues) > 1 {
			whereNameAndOrStr = fmt.Sprintf("WHERE (%v OR %v)", whereName, strings.Join(whereOrs, " OR "))
			for _, v := range whereOrValues {
				allValues = append(allValues, v)
			}
		} else {
			whereNameAndOrStr = fmt.Sprintf("WHERE %v", whereName)
		}

		if len(whereAndValues) == 1 {
			whereAndStr = fmt.Sprintf("AND %v", whereAnds[0])
			allValues = append(allValues, whereAndValues[0])
		} else if len(whereAndValues) > 1 {
			whereAndStr = fmt.Sprintf("AND %v", strings.Join(whereAnds, " AND "))
			for _, v := range whereAndValues {
				allValues = append(allValues, v)
			}
		} else {
			whereAndStr = ""
		}

		whereStr = fmt.Sprintf("%v %v", whereNameAndOrStr, whereAndStr)
		return whereStr, allValues
	}

}

//取得多条记录
func (orm *Orm) FindList(o interface{}) error {
	err := orm.scanStructIntoMap(o)
	if err != nil {
		return err
	}

	args := orm.StructMap

	//FILTER过滤不需要的字段
	fs := orm.FilterStrs
	var selectStr string
	if len(fs) < 1 {
		selectStr = "*"
	} else {
		for i := 0; i < len(fs); i++ {
			delete(args, fs[i])
		}

		//SELECT拼接条件
		var selectStrs []string
		for k, _ := range args {
			selectStrs = append(selectStrs, fmt.Sprintf("_%v", strings.ToLower(k)))
		}

		selectStr = strings.Join(selectStrs, ",")

	}

	fmt.Printf("selectStr=%v\n", selectStr)

	//WHERE STRING
	whereStr, paramValue := getWhereStr(orm.WhereMap)
	fmt.Printf("whereStr=%v\n", whereStr)

	//LIMIT OFFSET
	j := len(paramValue)
	var limitAndOffsetStr string
	if orm.LimitStr > 0 && orm.OffsetStr > 0 {
		var limit, offset interface{}
		limit = orm.LimitStr
		offset = orm.OffsetStr
		paramValue = append(paramValue, limit)
		paramValue = append(paramValue, offset)
		limitAndOffsetStr = fmt.Sprintf("LIMIT $%v OFFSET $%v", j+1, j+2)
	} else {
		limitAndOffsetStr = ""
	}

	//ORDER BY
	orderByStr := orm.OrderByStr

	orm.SqlStr = fmt.Sprintf("SELECT %v FROM %v %v %v %v", selectStr, orm.TableName, whereStr, orderByStr, limitAndOffsetStr)
	//orm.ParamValue = whereValues
	fmt.Printf("map=%v\n", orm.WhereMap)
	fmt.Printf("SQL=%v\n", orm.SqlStr)
	fmt.Printf("SQL VALUES=%v\n", paramValue)

	results, err := orm.query(orm.SqlStr, paramValue)
	if err != nil {
		return err
	}

	fmt.Printf("results=%v\n", results)
	fmt.Printf("results number=%v\n", len(results))
	/*
		structValue := reflect.Indirect(reflect.ValueOf(o))
		structValueType := structValue.Type().Elem()
		for _, v := range results {
			newValue := reflect.New(structValueType)
			err = orm.scanMapIntoStruct(newValue.Interface(), v)
			if err != nil {
				return err
			}
			structValue.Set(reflect.Append(structValue, reflect.Indirect(reflect.ValueOf(newValue.Interface()))))
		}
	*/

	return nil
}
