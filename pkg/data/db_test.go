package data_test

import (
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kmgreen2/agglo/pkg/config"
	"github.com/kmgreen2/agglo/pkg/data"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestModel is a test table used to illustrate the use of Database and TestDatabase
type TestModel struct {
	Foo int
	Bar string
}

// Model is a simple wrapper around the Database interface, which exposes the underlying database connection
type Model struct {
	db data.Database
}

func (model *Model) SetFooBar(foo int, bar string) error {
	gdb := model.db.Get()
	return gdb.Create(&TestModel{
		Foo: foo,
		Bar: bar,
	}).Error
}

func (model *Model) GetBarFromFoo(foo int) (string, error) {
	var testModel TestModel
	gdb := model.db.Get()
	err := gdb.Where("foo = ?", foo).First(&testModel).Error
	if err != nil {
		return "", err
	}
	return testModel.Bar, nil
}

func (model *Model) GetFooBars() ([]*TestModel, error) {
	testModel := make([]*TestModel, 0)
	gdb := model.db.Get()
	err := gdb.Find(&testModel).Error
	if err != nil {
		return nil, err
	}
	return testModel, nil
}

func TestNewDatabase(t *testing.T) {
	configBase, err := config.NewConfigBase()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	dbConfig := &data.DatabaseConfig{
		ConfigBase: configBase,
	}
	dbConfig.SetDatabaseType(data.MockDatabase)
	dbConfig.SetDSN("sqlmock_db_0")

	db, err := data.NewDatabaseImpl(dbConfig)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer db.Close()

	err = db.Ping()
	assert.Nil(t, err)
	gdb := db.Get()
	assert.NotNil(t, gdb)
}

func TestUseDatabase(t *testing.T) {
	configBase, err := config.NewConfigBase()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	dbConfig := &data.DatabaseConfig{
		ConfigBase: configBase,
	}
	dbConfig.SetDatabaseType(data.MockDatabase)
	dbConfig.SetDSN("sqlmock_db_0")

	db, err := data.NewDatabaseImpl(dbConfig)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer db.Close()

	dbModel := &Model {
		db,
	}

	mock := db.Mock()

	for i := 0; i < 10; i++ {
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO").
			WithArgs(i, fmt.Sprintf("%d", i)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		err = dbModel.SetFooBar(i, fmt.Sprintf("%d", i))
		if err != nil {
			assert.FailNow(t, err.Error())
		}
	}

	for i := 0; i < 10; i++ {
		rows := sqlmock.NewRows([]string{"foo", "bar"}).
			AddRow(i, fmt.Sprintf("%d", i))
		mock.ExpectQuery("SELECT (.+) FROM \"test_models\"").
			WithArgs(i).
			WillReturnRows(rows)
		bar, err := dbModel.GetBarFromFoo(i)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.Equal(t, fmt.Sprintf("%d", i), bar)
	}

	rows := sqlmock.NewRows([]string{"foo", "bar"})
	for i := 0; i < 10; i++ {
		rows.AddRow(i, fmt.Sprintf("%d", i))
	}
	mock.ExpectQuery("SELECT (.+) FROM \"test_models\"").
		WithArgs().
		WillReturnRows(rows)
	fooBars, err := dbModel.GetFooBars()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	for _, fooBar := range fooBars {
		assert.Equal(t, fmt.Sprintf("%d", fooBar.Foo), fooBar.Bar)
	}
}