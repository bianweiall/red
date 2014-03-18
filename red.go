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
	SelectStr string
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
	//RETURNING
	ReturningStr string
}

func (orm *Orm) InitOrm() {
	orm.SqlStr = ""
	orm.TableName = ""
	orm.ParamValues = make([]interface{}, 0)
	orm.PKName = ""
	orm.AutoPKName = ""
	orm.DateTimeNames = make([]string, 0)
	orm.StructMap = make(map[string]interface{})
	orm.SelectStr = ""
	orm.OrderByStr = ""
	orm.LimitStr = 0
	orm.OffsetStr = 0
	orm.WhereStr = ""
	orm.ReturningStr = ""
}

//New一个Orm对象
func New(driverName, dataSourceName string) (error, *Orm) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return err, nil
	}
	return nil, &Orm{Db: db}
}

//设置数据库表名，参数tableName格式："_" + 英文字母小写(对应数据库中的表名，例如：“_bookinfo”)
func (orm *Orm) SetTableName(tableName string) *Orm {
	if tableName != "" {
		//正则匹配
		if regexp.MustCompile(`^_[a-z]+\b`).MatchString(tableName) == true {
			orm.TableName = tableName
		}
	}
	return orm
}

//设置需要选择的数据库字段,参数str的字段名格式："_" + 英文字母小写(对应数据库中的字段，例如：“_id”)
func (orm *Orm) Select(strs ...string) *Orm {
	if len(strs) > 0 {
		i := true
		for _, v := range strs {
			if v != "" {
				//正则检查是否匹配，如果不匹配则 i=false，并跳出循环
				if regexp.MustCompile(`^_[a-z]+\b`).MatchString(v) != true {
					i = false
					break
				}
			}
		}

		if i == true {
			if len(strs) == 1 {
				orm.SelectStr = fmt.Sprintf("SELECT %v", strs[0])
			} else {
				orm.SelectStr = fmt.Sprintf("SELECT %v", strings.Join(strs, ","))
			}
		}
	}
	return orm
}

//设置Where条件,str参数中数据库字段格式："_" + 英文字母小写，其他都用英文字母小写
func (orm *Orm) Where(str string, strValue ...interface{}) *Orm {
	if str != "" && strValue != nil {
		//所有英文字母小写
		str = strings.ToLower(str)
		//计算"?"有几个
		num := strings.Count(str, "?")

		if num > 0 {
			if num == len(strValue) {
				//定义一个匿名函数，作用是检查str参数中的字段名格式是否正确，返回true为正确
				fn := func(strs []string) bool {
					i := true
					//定义一个保存字段名的数组
					var newStrs []string
					//循环删除如下字符，把得到字段名加入数组
					for _, str := range strs {
						if str != "" {
							//删除前后空格
							var v = strings.TrimSpace(str)
							//如果字符串中有"and"，执行删除"and"操作
							if strings.Contains(v, "and") == true {
								v = strings.Trim(v, "and")
							}
							if strings.Contains(v, "or") == true {
								v = strings.Trim(v, "or")
							}
							if strings.Contains(v, "=") == true {
								v = strings.Trim(v, "=")
							}
							if strings.Contains(v, "<>") == true {
								v = strings.Trim(v, "<>")
							}
							if strings.Contains(v, "!=") == true {
								v = strings.Trim(v, "!=")
							}
							if strings.Contains(v, ">") == true {
								v = strings.Trim(v, ">")
							}
							if strings.Contains(v, "<") == true {
								v = strings.Trim(v, "<")
							}
							if strings.Contains(v, ">=") == true {
								v = strings.Trim(v, ">=")
							}
							if strings.Contains(v, "<=") == true {
								v = strings.Trim(v, "<=")
							}
							if strings.Contains(v, "between") == true {
								v = strings.Trim(v, "between")
							}
							if strings.Contains(v, "like") == true {
								v = strings.Trim(v, "like")
							}
							v = strings.TrimSpace(v)
							if v != "" {
								newStrs = append(newStrs, v)
							}
						}
					}
					//fmt.Println("newStrs: ", newStrs)
					//循环匹配字段名是否符合要求
					for _, v := range newStrs {
						//正则检查是否匹配，如果不匹配则 i=false，并跳出循环
						if regexp.MustCompile(`^_[a-z]+\b`).MatchString(v) != true {
							i = false
							break
						}
					}
					return i
				}
				//用"?"分割成字符串数组
				strs := strings.Split(str, "?")
				//用上面的匿名函数检查是否正确
				i := fn(strs)
				//fmt.Println("ok: ", i)
				if i == true {
					if num == 1 {
						//用"S1"替换"?"
						str = strings.Replace(str, "?", "$1", 1)
					} else {
						//循环用占位符替换"?"
						for i := 1; i <= num; i++ {
							str = strings.Replace(str, "?", fmt.Sprintf("$%v", i), 1)
						}
					}
					orm.WhereStr = fmt.Sprintf("WHERE %v", str)
					orm.WhereStrValue = strValue
				}
			}
		}
	}
	//fmt.Println("orm wherestr: ", orm.WhereStr)
	return orm
}

//设置ORDER BY，orderByStrs参数中字段名格式("_" + 英文字母小写)，其他都用英文字母小写
func (orm *Orm) OrderBy(orderByStrs ...string) *Orm {
	if len(orderByStrs) > 0 {
		//定义一个保存字段名的数组
		var newStrs []string
		//循环删除如下字符，把得到字段名加入数组
		for _, v := range orderByStrs {
			if v != "" {
				//所有英文字母小写
				str := strings.ToLower(v)
				//删除前后空格
				str = strings.TrimSpace(str)
				//如果字符串中有"asc"，执行删除"asc"操作
				if strings.Contains(str, "asc") == true {
					str = strings.Trim(str, "asc")
				}
				if strings.Contains(str, "desc") == true {
					str = strings.Trim(str, "desc")
				}
				str = strings.TrimSpace(str)
				newStrs = append(newStrs, str)
			}
		}
		i := true
		//循环匹配字段名是否符合要求
		for _, v := range newStrs {
			//正则检查是否匹配，如果不匹配则 i=false，并跳出循环
			if regexp.MustCompile(`^_[a-z]+\b`).MatchString(v) != true {
				i = false
				break
			}
		}
		if i == true {
			var orderByStr string
			if len(orderByStrs) == 1 {
				orderByStr = fmt.Sprintf("ORDER BY %v", orderByStrs[0])
			} else {
				orderByStr = fmt.Sprintf("ORDER BY %v", strings.Join(orderByStrs, ", "))
			}

			orm.OrderByStr = orderByStr
		}
	}
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

//设置RETURNING, 参数中字段名格式("_" + 英文字母小写)
func (orm *Orm) Returning(strs ...string) *Orm {
	if len(strs) > 0 {
		i := true
		//循环匹配字段名是否符合要求
		for _, v := range strs {
			//正则检查是否匹配，如果不匹配则 i=false，并跳出循环
			if regexp.MustCompile(`^_[a-z]+\b`).MatchString(v) != true {
				i = false
				break
			}
		}
		if i == true {
			var returningStr string
			if len(strs) == 1 {
				returningStr = fmt.Sprintf("RETURNING %v", strs[0])
			} else {
				returningStr = fmt.Sprintf("RETURNING %v", strings.Join(strs, ", "))
			}

			orm.ReturningStr = returningStr
		}
	}
	return orm
}

//传入一个struct对象，返回错误和两个字符串，此两个字符串用于拼接SQL字符串的
//例如有一个SQL语句:INSERT INTO table_name (_id,_name,_age) VALUES ($1,$2,$3)
//返回的第一个string代表字段名字符串(类似上面“_id,_name,_age”处)，返回的第二个string代表字段值占位符字符串(类似上面“$1,$2,$3”处)
func (orm *Orm) getInsert(o interface{}) error {
	//把一个struct类型的信息保存到一个Orm中
	err := orm.scanStructIntoOrm(o)
	if err != nil {
		return err
	}
	//获取刚刚保存的StructMap
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
	//合成INSERT语句
	orm.SqlStr = fmt.Sprintf("INSERT INTO %v(%v) VALUES(%v)", orm.TableName, insertFieldStr, insertValueStr)
	return nil
}

//保存数据,有返回值并返回一个map
func (orm *Orm) Create(o interface{}) (error, map[string][]byte) {
	//关闭资源
	defer orm.InitOrm()
	//取得需要保存到数据库中的字段名字符串和字段值占位符字符串
	err := orm.getInsert(o)
	if err != nil {
		return err, nil
	}

	//判断是否有RETURNING语句
	if orm.ReturningStr == "" {
		//持久化到数据库
		_, err = orm.exec()

		if err != nil {
			return err, nil
		}
		return nil, nil
	} else {
		//合成INSERT ... RETURNING
		sql := fmt.Sprintf("%v %v", orm.SqlStr, orm.ReturningStr)

		resultsSlice, err := orm.query(sql, orm.ParamValues)
		if err != nil {
			return err, nil
		}
		return nil, resultsSlice[0]
	}

}

//更新数据
func (orm *Orm) Update(o interface{}) error {
	defer orm.InitOrm()
	//把STRUCT相关信息映射到ORM
	err := orm.scanStructIntoOrm(o)
	if err != nil {
		return err
	}
	//获取STRUCT字段名和值
	args := orm.StructMap
	//删除带`auto:pk`的字段
	if orm.AutoPKName != "" {
		delete(args, orm.AutoPKName)
	}

	//保存字段名的数组
	var names []string
	//保存字段值占位符的数组
	var namevlues []string
	//保存字段值的数组
	var values []interface{}

	var snum = 1
	//循环提取内容到相关的数组
	for k, v := range args {
		names = append(names, k)
		if v == "current_timestamp" {
			namevlues = append(namevlues, "current_timestamp")
		} else {
			namevlues = append(namevlues, fmt.Sprintf("$%v", snum))
			values = append(values, v)
			snum++
		}
	}
	//fmt.Println("names: ", names)
	//fmt.Println("namevlues: ", namevlues)
	//fmt.Println("values: ", values)
	whereStr, whereValue := orm.WhereStr, orm.WhereStrValue
	//检查WHERE条件是否正确
	if whereStr == "" {
		return errors.New("where条件不能为空，请检查Where()参数是否正确")
	}
	//fmt.Println("whereStr: ", whereStr)
	//fmt.Println("whereValue: ", whereValue)
	//WHERE SQL
	for i := 1; i <= len(orm.WhereStrValue); i++ {
		whereStr = strings.Replace(whereStr, fmt.Sprintf("$%v", i), fmt.Sprintf("$%v", snum), 1)
		snum = snum + i
	}
	values = append(values, whereValue)
	fmt.Println("whereStr2: ", whereStr)
	var setStrs []string
	for i := 0; i < len(names); i++ {
		setStrs = append(setStrs, fmt.Sprintf("%v=%v", fmt.Sprintf("_%v", strings.ToLower(names[i])), namevlues[i]))
	}
	//SET SQL
	setStr := strings.Join(setStrs, ",")
	fmt.Println("setStr: ", setStr)

	orm.ParamValues = values
	orm.SqlStr = fmt.Sprintf("UPDATE %v SET %v %v", orm.TableName, setStr, whereStr)
	fmt.Println("orm.SqlStr: ", orm.SqlStr)
	_, err = orm.exec()
	if err != nil {
		return err
	}
	return nil
}

//删除数据
func (orm *Orm) Delete(o interface{}) error {
	defer orm.InitOrm()
	err := orm.scanStructIntoOrm(o)
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
		err := orm.scanStructIntoOrm(st.Interface())
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
	err := orm.scanStructIntoOrm(o) //把一个struct信息保存到一个Orm中
	if err != nil {
		return err
	}

	//args := orm.StructMap //获取上面保存的map
	//fs := orm.SelectStr   //获取需要过滤掉的字段

	//for i := 0; i < len(fs); i++ {
	//	delete(args, fs[i]) //从map中删除需要过滤的字段
	//}

	//var selectStrs []string //定义一个保存所有SELECT字段的数组
	//for k, _ := range args {
	//	selectStrs = append(selectStrs, fmt.Sprintf("_%v", strings.ToLower(k))) //每个字段所有字母全部小写，并加上“_”前缀，对应数据库中的字段
	//}

	//selectStr := strings.Join(selectStrs, ",") //合成SELECT字符串

	whereStrName, whereValue := orm.WhereStr, orm.WhereStrValue //获取SQL WHERE语句和WHERE语句中所有？代表的值的集合
	if whereStrName == "" {
		return errors.New("where()参数输入错误")
	}

	orm.SqlStr = fmt.Sprintf("SELECT %v FROM %v WHERE %v", orm.SelectStr, orm.TableName, whereStrName) //合成SQL语句

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
	//args := orm.StructMap
	//fs := orm.SelectStr
	//var selectStr string
	//if len(fs) < 1 {
	//	selectStr = "*"
	//} else {
	//	for i := 0; i < len(fs); i++ {
	//		delete(args, fs[i])
	//	}
	//	var selectStrs []string
	//	for k, _ := range args {
	//		selectStrs = append(selectStrs, fmt.Sprintf("_%v", strings.ToLower(k)))
	//	}
	//	selectStr = strings.Join(selectStrs, ",")
	//}

	selectSql := fmt.Sprintf("SELECT %v FROM %v %v %v %v %v", orm.SelectStr, orm.TableName, whereStr, orm.OrderByStr, limitStr, offsetStr)
	orm.SqlStr = selectSql

	return selectSql, whereStrValue

}

//解析STRUCT字段和值到一个MAP中
func (orm *Orm) scanStructIntoOrm(o interface{}) error {
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
	//var dateTimes []string
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
			args[t.Field(i).Name] = "current_timestamp"
			//dateTimes = append(dateTimes, t.Field(i).Name)
		}

		if j > 1 {
			return errors.New("要求struct字段只能设置一个主键")
		}
	}

	if orm.TableName == "" {
		orm.TableName = fmt.Sprintf("_%v", strings.ToLower(t.Name()))
	}

	//orm.DateTimeNames = dateTimes
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
