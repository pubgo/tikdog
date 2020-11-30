package xorm

import (
	"fmt"
	"github.com/pubgo/xerror"
	"testing"
)
import "xorm.io/xorm"
import _ "github.com/mattn/go-sqlite3"

func TestName(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3",  "file:ent?mode=memory&cache=shared&_fk=1")
	xerror.Panic(err)
	fmt.Println(engine.DB().Ping())
}
