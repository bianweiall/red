RED
====

RED是一个GO语言的ORM，参考了beedb，主要是用来学习GO语言的，只支持postgresql数据库，支持Create,Update,Delete,Find这几个方法。

可以自定义SQL语句中的SELECT / (WHERE OR AND) / ORDER BY / LIMIT OFFSET / RETURNING语句。

安装：
```go
go get github.com/bianweiall/red
```

用法：
数据库表和字段一律"_" + 英文字母小写
定义结构，对应数据库表结构(例如：User=>_user, Id=>_id, Name=>_name, Time=>_time, Age=>_age)
```go
type User struct {
    Id   int `pk:auto` \\`pk:auto`表示是自动递增主键，`pk`为一般主键
    Name string
	Time string `dt`   \\`dt`表示postgresql中的日期时间类型"current_timestamp"
	Age  int
}
```

取得Orm
```go
err, orm := red.New()
```

```go
func main() {
    var user User
	err, orm := red.New()
	if err != nil {
		panic(err)
	}
	err = orm.Where("_id=?", 1).Find(&user)
	if err != nil {
		panic(err)
	}
}
```

保存，不用给定自增Id和Time，数据库会自动插入，如果设置了Returning则返回err和map，如果没设置Returning则返回err和nil
```go
var user User
user.Name="小明"
user.Age=12
err, r := orm.Create(&user)
```

保存，并返回刚插入的这条记录的Id和Name
```go
var user User
user.Name="小明"
user.Age=12
err, strs := orm.Returning("_id", "_name").Create(&user)
```

更新，默认更新所有字段(自增主键除外)
```go
var user User
user.Name="鸭蛋"
user.Age=50
err := orm.Where("_id = ?", 1).Update(&user)
```

更新指定字段
```go
var user User
user.Name="鸭蛋"
user.Age=50
err := orm.Set("_time", "_name", "_age").Where("_id = ?", 1).Update(&user)
```

删除一个User
```go
var user User
err = orm.Where("_id = ?", 1).Delete(&user)
```

查询一个User，不设置select默认取得所有字段
```go
var user User
err := orm.Select("_name", "_age").Where("_id = ?", 1).Find(&user)
```

查询User集合，不设置select默认取得所有字段
```go
var users []User
err := orm.Where("_id > ? and _name like ?", 1, "%关键字%").Find(&user)
```

一以下方法传递进来的参数中数据库字段的格式："_" + 英文字母小写
不设置默认是："_" + 对象英文字母小写，例如："_user"
func (orm *Orm) SetTableName(tableName string) *Orm
```go
orm := orm.SetTableName("_book")
```

设置选择的字段，不设置默认为选择全部字段
func (orm *Orm) Select(strs ...string) *Orm
```go
orm := orm.Select("_id","_name")
```

UPDATE语句的SET后面跟着的表达式，不设置默认为全部字段都提交（自增主键除外）
func (orm *Orm) Set(strs ...string) *Orm
```go
orm := orm.Set("_age","_name")
```

WHERE语句后面跟着的表达式
func (orm *Orm) Where(str string, strValue ...interface{}) *Orm
```go
orm := orm.Where("_id > ? and _name like ?", 1, "%关键字%")
```

LIMIT和OFFSET，参数为大于0的整数
func (orm *Orm) Limit(limit int) *Orm
func (orm *Orm) Offset(offset int) *Orm
```go
err = orm.Where("_id > ?", 1).Limit(10).Offset(5)
```

ORDER BY 
func (orm *Orm) OrderBy(orderByStrs ...string) *Orm
```go
err := orm.Where("_id > ?", 1).OrderBy("_id desc","_age asc").Find(&users)
```