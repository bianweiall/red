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

func TestSetTableName(t *testing.T) {
	err, orm := New("postgres", "user=greenerp password=guotinghuayuan30301 dbname=greenerp sslmode=disable")
	if err != nil {
		fmt.Println(err)
	}
	orm.SetTableName("_book")
	fmt.Println(orm.TableName)
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
