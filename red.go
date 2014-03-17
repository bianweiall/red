package red

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lxn/go-pgsql"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Orm struct {
	Db *sql.DB
	//需要执行的SQL语句
	SqlStr string
	//数据库表名
	TableName string
	//SQL中参数的值
	ParamValues []interface{}
	//数据库主键，tag是`pk:auto`为自动递增主键，`pk`为不能自动递增主键
	PKName string
	//自动主键
	AutoPKName string
	//tag为`dt`的字段名,表示日期和时间
	DateTimeNames []string
	//解析struct为一个map
	StructMap map[string]interface{}
	//过滤字符串
	FilterStrs []string
	//ORDER BY字符串
	OrderByStr string
	//LIMIT字符串
	LimitStr int
	//OFFSET字符串
	OffsetStr int
	//WHERE字符串
	WhereStr string
	//WHERE字符串中值
	WhereStrValue []interface{}
}

func (orm *Orm) InitOrm() {
	orm.SqlStr = ""
	orm.TableName = ""
	orm.ParamValues = make([]interface{}, 0)
	orm.PKName = ""
	orm.AutoPKName = ""
	orm.DateTimeNames = make([]string, 0)
	orm.StructMap = make(map[string]interface{})
	orm.FilterStrs = make([]string, 0)
	orm.OrderByStr = ""
	orm.LimitStr = 0
	orm.OffsetStr = 0
	orm.WhereStr = ""
}

func New(driverName, dataSourceName string) (error, *Orm) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return err, nil
	}
	return nil, &Orm{Db: db}
}

//设置数据库表名，参数tableName格式："_" + 英文字母小写(例如：“_bookinfo”)
func (orm *Orm) SetTableName(tableName string) *Orm {
	if tableName != "" {
		//正则匹配
		if regexp.MustCompile(`^_[a-z]+\b`).MatchString(tableName) == true {
			orm.TableName = tableName
		}
	}
	return orm
}

//设置过滤条件
func (orm *Orm) Filter(filter string) *Orm {
	var strs []string
	if filter != "" {
		if strings.Contains(filter, ",") == true {
			strs = strings.Split(filter, ",")
			orm.FilterStrs = strs
		} else {
			orm.FilterStrs = append(strs, filter)
		}
	}
	return orm
}

//设置Where条件
func (orm *Orm) Where(str string, strValue ...interface{}) *Orm {
	if str != "" && strValue != nil {
		//orm.WhereStr = str
		//orm.WhereStrValue = strValue
		num := strings.Count(str, "?")
		if num > 0 {
			if num == len(strValue) {
				if num == 1 {
					strings.Replace(str, "?", "$1", 1)
				} else {
					for i := 1; i <= num; i++ {
						strings.Replace(str, "?", fmt.Sprintf("$%v", i), 1)
					}
				}
				orm.WhereStr = str
				orm.WhereStrValue = strValue
			}
		}
	}
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

//传入一个struct对象，返回错误和两个字符串，此两个字符串用于拼接SQL字符串的
//例如有一个SQL语句:INSERT INTO table_name (_id,_name,_age) VALUES ($1,$2,$3)
//返回的第一个string代表字段名字符串(类似上面“_id,_name,_age”处)，返回的第二个string代表字段值占位符字符串(类似上面“$1,$2,$3”处)
func (orm *Orm) getInsertStr(o interface{}) (error, string, string) {
	//把一个struct类型的值保存到一个map中
	err := orm.scanStructIntoMap(o)
	if err != nil {
		return err, "", ""
	}
	//获取刚刚保存的map
	args := orm.StructMap
	//获取表示时间和日期的字段
	dt := orm.DateTimeNames
	//删除表示自动主键的字段
	delete(args, orm.AutoPKName)
	//循环删除表示时间和日期的字段
	for j := 0; j < len(dt); j++ {
		delete(args, dt[j])
	}
	//定义保存字段名的数组
	var names []string
	//定义保存字段值占位符的字符串数组
	var namevlues []string
	//定义保存字段值的数组
	var values []interface{}
	//循环把map中的字段名和值保存到数组中
	for n, v := range args {
		names = append(names, n)
		values = append(values, v)
	}

	//通过字段的个数把占位符添加到数组中
	for j := 1; j <= len(names); j++ {
		namevlues = append(namevlues, fmt.Sprintf("$%v", j))
	}

	//将表示时间和日期的字段添加到数组
	//将"current_timestamp"添加到占位符数组
	for j := 0; j < len(dt); j++ {
		names = append(names, dt[j])
		namevlues = append(namevlues, "current_timestamp")
	}
	//合成插入字段字符串（类似"_id,_name,_age"）
	insertFieldStr := fmt.Sprintf("_%v", strings.ToLower(strings.Join(names, ",_")))
	//合成字段值占位符字符串（类似"$1,$2,$3,current_timestamp"）
	insertValueStr := strings.Join(namevlues, ",")
	//保存字段值到orm
	orm.ParamValues = values

	return nil, insertFieldStr, insertValueStr

}

//保存数据
func (orm *Orm) Create(o interface{}) error {
	//关闭资源
	defer orm.InitOrm()
	//取得需要保存到数据库中的字段名字符串和字段值占位符字符串
	err, tableItemStr, valueItemStr := orm.getInsertStr(o)
	if err != nil {
		return err
	}
	//合成INSERT语句
	orm.SqlStr = fmt.Sprintf("INSERT INTO %v(%v) VALUES(%v)", orm.TableName, tableItemStr, valueItemStr)
	//持久化到数据库
	_, err = orm.exec()

	if err != nil {
		return err
	}
	return nil
}

//保存数据后返回自增ID
func (orm *Orm) CreateAndReturnId(o interface{}) (error, int) {
	//关闭资源
	defer orm.InitOrm()
	//取得需要保存到数据库中的字段名字符串和字段值占位符字符串
	err, tableItemStr, valueItemStr := orm.getInsertStr(o)
	if err != nil {
		return err, 0
	}
	//合成INSERT语句，需要返回这条记录的ID
	sql := fmt.Sprintf("INSERT INTO %v(%v) VALUES(%v) RETURNING _id", orm.TableName, tableItemStr, valueItemStr)
	//执行Query方法
	rows, err := orm.Db.Query(sql, orm.ParamValues...)
	if err != nil {
		return err, 0
	}

	var id int
	for rows.Next() {
		//循环得到id值
		err := rows.Scan(&id)
		if err != nil {
			return err, 0
		}
	}
	return nil, id
}

//更新数据
func (orm *Orm) Update(o interface{}) error {
	defer orm.InitOrm()
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

	_, err = orm.exec()
	if err != nil {
		return err
	}
	return nil
}

//删除数据
func (orm *Orm) Delete(o interface{}) error {
	defer orm.InitOrm()
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

	_, err = orm.exec()
	if err != nil {
		return err
	}
	return nil
}

//取得1条记录或多条记录
func (orm *Orm) Find(slicePtr interface{}) error {
	defer orm.InitOrm()
	sliceValue := reflect.Indirect(reflect.ValueOf(slicePtr))
	if sliceValue.Kind() == reflect.Struct { //取得一条记录
		err := orm.findOne(slicePtr)
		if err != nil {
			return err
		}
	} else { //取得多条记录
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
			err = orm.scanMapIntoStruct(newValue.Interface(), results)
			if err != nil {
				return err
			}
			sliceValue.Set(reflect.Append(sliceValue, reflect.Indirect(reflect.ValueOf(newValue.Interface()))))
		}
	}

	return nil
}

//取得一条记录
func (orm *Orm) findOne(o interface{}) error {
	//defer orm.InitOrm()
	err := orm.scanStructIntoMap(o) //把一个struct信息保存到一个map中
	if err != nil {
		return err
	}

	args := orm.StructMap //获取上面保存的map
	fs := orm.FilterStrs  //获取需要过滤掉的字段

	for i := 0; i < len(fs); i++ {
		delete(args, fs[i]) //从map中删除需要过滤的字段
	}

	var selectStrs []string //定义一个保存所有SELECT字段的数组
	for k, _ := range args {
		selectStrs = append(selectStrs, fmt.Sprintf("_%v", strings.ToLower(k))) //每个字段所有字母全部小写，并加上“_”前缀，对应数据库中的字段
	}

	selectStr := strings.Join(selectStrs, ",") //合成SELECT字符串

	whereStrName, whereValue := orm.WhereStr, orm.WhereStrValue //获取SQL WHERE语句和WHERE语句中所有？代表的值的集合
	if whereStrName == "" {
		return errors.New("where()参数输入错误")
	}

	orm.SqlStr = fmt.Sprintf("SELECT %v FROM %v WHERE %v", selectStr, orm.TableName, whereStrName) //合成SQL语句

	results, err := orm.query(orm.SqlStr, whereValue) //执行query，返回一个map数组
	if err != nil {
		return err
	}

	err = orm.scanMapIntoStruct(o, results[0]) //把上面的MAP映射到一个struct中
	if err != nil {
		return err
	}

	return nil
}

func (orm *Orm) query(sqlStr string, args []interface{}) ([]map[string][]byte, error) {
	rows, err := orm.Db.Query(sqlStr, args...)
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

//持久化到数据库
func (orm *Orm) exec() (sql.Result, error) {
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

func (orm *Orm) getSelectSqlAndValues() (string, []interface{}) {
	//var allValues []interface{}
	//WHERE
	whereStr, whereStrValue := orm.WhereStr, orm.WhereStrValue
	var j int
	if whereStr == "" {
		j = 0
	} else {
		j = len(whereStrValue)
	}

	//LIMIT OFFSET
	var limitStr, offsetStr string
	if orm.LimitStr < 1 {
		limitStr = ""
	} else {
		limitStr = fmt.Sprintf("LIMIT $%v", j+1)
		var l interface{}
		l = orm.LimitStr
		//allValues = append(allValues, l)
		whereStrValue = append(whereStrValue, l)
		j++
	}

	if orm.OffsetStr < 1 {
		offsetStr = ""
	} else {
		offsetStr = fmt.Sprintf("OFFSET $%v", j+1)
		var o interface{}
		o = orm.OffsetStr
		//allValues = append(allValues, o)
		whereStrValue = append(whereStrValue, o)
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

	selectSql := fmt.Sprintf("SELECT %v FROM %v %v %v %v %v", selectStr, orm.TableName, whereStr, orm.OrderByStr, limitStr, offsetStr)
	orm.SqlStr = selectSql

	return selectSql, whereStrValue

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
