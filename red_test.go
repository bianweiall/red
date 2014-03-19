package red

import (
	"fmt"
	//"strings"
	"testing"
)

type Warehouse struct {
	Id    int `pk:auto`
	Name  string
	Level int
	Fid   int
}

type Book struct {
	Id      int `pk:auto`
	Name    string
	Level   int
	Fid     int
	OldTime string `dt`
	NowTime string `dt`
}

//func TestSetTableName(t *testing.T) {
//	err, orm := New("postgres", "user=greenerp password=guotinghuayuan30301 dbname=greenerp sslmode=disable")
//	if err != nil {
//		fmt.Println(err)
//	}
//	orm.SetTableName("_book")
//	fmt.Println("_book: ", orm.TableName)
//	orm.TableName = ""
//	orm.SetTableName("_book111")
//	fmt.Println("_book111: ", orm.TableName)
//	orm.TableName = ""
//	orm.SetTableName("book")
//	fmt.Println("book: ", orm.TableName)
//	orm.TableName = ""
//	orm.SetTableName("_bookinfo")
//	fmt.Println("_bookinfo: ", orm.TableName)
//}

//func TestSelect(t *testing.T) {
//	err, orm := New("postgres", "user=greenerp password=guotinghuayuan30301 dbname=greenerp sslmode=disable")
//	if err != nil {
//		fmt.Println(err)
//	}
//	orm.Select("_id")
//	fmt.Println("_id: ", orm.SelectStr)

//	orm.SelectStr = ""
//	orm.Select("_id111")
//	fmt.Println("_id111: ", orm.SelectStr)

//	orm.SelectStr = ""
//	orm.Select("_id", "_name")
//	fmt.Println("_id ,_name: ", orm.SelectStr)
//}

//func TestWhere(t *testing.T) {
//	err, orm := New("postgres", "user=greenerp password=guotinghuayuan30301 dbname=greenerp sslmode=disable")
//	if err != nil {
//		fmt.Println(err)
//	}
//	orm.Where("_id=? and _name=? or _age=? and _name like ?", 1, "刘德华", 50, "%看看%")
//	fmt.Println("where: ", orm.WhereStr)

//}

//func TestOrderBy(t *testing.T) {
//	err, orm := New("postgres", "user=greenerp password=guotinghuayuan30301 dbname=greenerp sslmode=disable")
//	if err != nil {
//		fmt.Println(err)
//	}
//	orm.OrderBy("_id desc")
//	fmt.Println("order by str: ", orm.OrderByStr)

//}

//func TestLimit(t *testing.T) {
//	err, orm := New("postgres", "user=greenerp password=guotinghuayuan30301 dbname=greenerp sslmode=disable")
//	if err != nil {
//		fmt.Println(err)
//	}
//	orm.Limit(20)
//	fmt.Println("LimitStr: ", orm.LimitStr)

//}

//func TestReturning(t *testing.T) {
//	err, orm := New("postgres", "user=greenerp password=guotinghuayuan30301 dbname=greenerp sslmode=disable")
//	if err != nil {
//		fmt.Println(err)
//	}

//	orm.Returning("_id", "_name")
//	fmt.Println("ReturningStr: ", orm.ReturningStr)

//}

//func TestGetInsertStr(t *testing.T) {
//	err, orm := New("postgres", "user=greenerp password=guotinghuayuan30301 dbname=greenerp sslmode=disable")
//	if err != nil {
//		fmt.Println(err)
//	}

//	var b Book
//	b.Fid = 0
//	b.Level = 0
//	b.Name = "未定义1"

//	err = orm.getInsert(&b)
//	if err != nil {
//		fmt.Println(err)
//	}
//	fmt.Println("insert sql str: ", orm.SqlStr)
//	fmt.Println("insert sql value: ", orm.ParamValues)

//}

//func TestCreate(t *testing.T) {
//	err, orm := New("postgres", "user=greenerp password=guotinghuayuan30301 dbname=greenerp sslmode=disable")
//	if err != nil {
//		fmt.Println(err)
//	}

//	var w Warehouse
//	w.Fid = 0
//	w.Level = 0
//	w.Name = "未定义def"
//	err, strs := orm.Returning("_id", "_name").Create(&w)
//	if err != nil {
//		fmt.Println(err)
//	}
//	fmt.Println("_id: ", string(strs["_id"]))
//	fmt.Println("_name: ", string(strs["_name"]))
//}

func TestUpdate(t *testing.T) {
	err, orm := New("postgres", "user=greenerp password=guotinghuayuan30301 dbname=greenerp sslmode=disable")
	if err != nil {
		fmt.Println(err)
	}

	var w Warehouse
	w.Fid = 0
	//w.Level = 1
	w.Name = "未定义"
	err = orm.Where("_id = ?", 145).Update(&w)
	if err != nil {
		fmt.Println(err)
	}
}

//func TestDelete(t *testing.T) {
//	err, orm := New("postgres", "user=greenerp password=guotinghuayuan30301 dbname=greenerp sslmode=disable")
//	if err != nil {
//		fmt.Println(err)
//	}

//	var w Warehouse
//	err = orm.Where("_id = ?", 132).Delete(&w)
//	if err != nil {
//		fmt.Println(err)
//	}
//}
