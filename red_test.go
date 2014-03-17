package red

import (
	"fmt"
	"testing"
)

type Warehouse struct {
	Id    int `pk:auto`
	Name  string
	Level int
	Fid   int
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
//	fmt.Println("_id: ", orm.SelectStrs)

//	orm.SelectStrs[0] = ""
//	orm.Select("_id111")
//	fmt.Println("_id111: ", orm.SelectStrs)

//	orm.SelectStrs[0] = ""
//	orm.Select("_id,_name")
//	fmt.Println("_id ,_name: ", orm.SelectStrs)
//}

func TestWhere(t *testing.T) {
	err, orm := New("postgres", "user=greenerp password=guotinghuayuan30301 dbname=greenerp sslmode=disable")
	if err != nil {
		fmt.Println(err)
	}
	orm.Where("_id=? and name=? or age=? and name like ?", 1, "刘德华", 50, "%看看%")
	fmt.Println("where: ", orm.WhereStr)

}

//func TestCreate(t *testing.T) {
//	err, orm := New("postgres", "user=greenerp password=guotinghuayuan30301 dbname=greenerp sslmode=disable")
//	if err != nil {
//		fmt.Println(err)
//	}

//	var w Warehouse
//	w.Fid = 0
//	w.Level = 0
//	w.Name = "未定义1"
//	err = orm.Create(&w)
//	if err != nil {
//		fmt.Println(err)
//	}
//}

//func TestCreateAndReturnId(t *testing.T) {
//	err, orm := New("postgres", "user=greenerp password=guotinghuayuan30301 dbname=greenerp sslmode=disable")
//	if err != nil {
//		fmt.Println(err)
//	}

//	var w Warehouse
//	w.Fid = 0
//	w.Level = 0
//	w.Name = "未定义"
//	err, id := orm.CreateAndReturnId(&w)
//	if err != nil {
//		fmt.Println(err)
//	}
//	fmt.Println(id)
//}
