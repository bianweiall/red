RED
====

RED是一个GO语言的ORM，参考了beedb，主要是用来学习GO语言的，只支持postgresql数据库，支持Create,Update,Delete,FindOne,FindList这几个方法。

支持Filter过滤数据库字段，支持SQL (WHERE OR AND) / ORDER BY / LIMIT OFFSET 语句。

安装：
```go
go get github.com/bianweiall/red
```

用法：

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
	err = orm.Filter("Time").Where("Id", 1).FindOne(&user)
	if err != nil {
		panic(err)
	}
}
```

保存一个User到数据库，不用给定Id和Time，数据库会自动插入
```go
var user User
user.Name="小明"
user.Age=12
err = orm.Create(&user)
```

更新一个User到数据库
```go
var user User
user.Name="鸭蛋"
user.Age=50
orm.Where("Id", 1).Update(&user)
```

通过Id删除一个User
```go
var user User
err = orm.Where("Id", 2).Delete(&user)
```

通过Id查询一个User
```go
var user User
err = orm.Where("Id", 2).FindOne(&user)
```
或者是
```go
var user User
err = orm.Where("Id", 2).Find(&user)
```

如果不想查询Time这个字段
```go
var user User
err = orm.Filter("Time").Where("Id", 2).FindOne(&user)
```

查询User所有记录
```go
var users []User
err = orm.Find(&users)
```
使用WHERE条件查询,并且过滤不需要查询的字段
```go
err = orm.Filter("Time").Where("Age", 50).Find(&users)
```
要排序有"-"表示DESC，没有表示ASC
```go
err = orm.Filter("Time").Where("Age", 50).OrderBy("-Id","Age").Find(&users)
```
LIMIT OFFSET
```go
err = orm.Filter("Time").Where("Age", 50).OrderBy("-Id","Age").Limit(10).Offset(5).Find(&users)
```
WHERE OR 相当于SQL: WHERE (_age=50 OR _id=5 OR _id=8)
```go
err = orm.Filter("Time").Where("Age", 50).WhereOr("Id,Id",5,8).OrderBy("-Id","Age").Limit(10).Offset(5).Find(&users)
```
WHERE AND 相当于SQL: WHERE _age=50 AND _id=5 AND _name='鸭蛋'
```go
err = orm.Filter("Time").Where("Age", 50).WhereAnd("Id,Name",5,"鸭蛋").OrderBy("-Id","Age").Limit(10).Offset(5).Find(&users)
```
WHERE OR AND 相当于SQL: WHERE (_age=50 OR _name='鸭蛋') AND _id=5
```go
err = orm.Filter("Time").Where("Age", 50).WhereOr("Name","鸭蛋").WhereAnd("Id",5).OrderBy("-Id","Age").Limit(10).Offset(5).Find(&users)
```
