/*
 * Copyright (C) 2016-2018. ActionTech.
 * Based on: github.com/actiontech/dtle, github.com/github/gh-ost .
 * License: MPL version 2: https://www.mozilla.org/en-US/MPL/2.0 .
 */

package base

import (
	gosql "database/sql"
	"reflect"
	"testing"
	"time"

	umconf "github.com/actiontech/dtle/olddtle/internalinternal/config/mysql"

	"fmt"

	"github.com/actiontech/dtle/olddtle/internalinternal/client/driver/mysql/sql"

	test "github.com/outbrain/golib/tests"
	gomysql "github.com/siddontang/go-mysql/mysql"
)

func TestStringContainsAll(t *testing.T) {
	s := `insert,delete,update`

	test.S(t).ExpectFalse(StringContainsAll(s))
	test.S(t).ExpectFalse(StringContainsAll(s, ""))
	test.S(t).ExpectFalse(StringContainsAll(s, "drop"))
	test.S(t).ExpectTrue(StringContainsAll(s, "insert"))
	test.S(t).ExpectFalse(StringContainsAll(s, "insert", "drop"))
	test.S(t).ExpectTrue(StringContainsAll(s, "insert", ""))
	test.S(t).ExpectTrue(StringContainsAll(s, "insert", "update", "delete"))
}

func TestPrettifyDurationOutput(t *testing.T) {
	type args struct {
		d time.Duration
	}
	tests := []struct {
		name string
		args args
		want string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PrettifyDurationOutput(tt.args.d); got != tt.want {
				t.Errorf("PrettifyDurationOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSelfBinlogCoordinates(t *testing.T) {
	type args struct {
		db *gosql.DB
	}
	db, err := sql.CreateDB(fmt.Sprintf("root:rootroot@tcp(192.168.99.100:13307)/?timeout=5s&tls=false&autocommit=true&charset=utf8mb4,utf8,latin1&multiStatements=true"))
	if err != nil {
		return
	}
	tests := []struct {
		name                      string
		args                      args
		wantSelfBinlogCoordinates *BinlogCoordinateTx
		wantErr                   bool
	}{
		// TODO: Add test cases.
		{"T1", args{db}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSelfBinlogCoordinates, err := GetSelfBinlogCoordinates(tt.args.db)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSelfBinlogCoordinates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotSelfBinlogCoordinates, tt.wantSelfBinlogCoordinates) {
				t.Errorf("GetSelfBinlogCoordinates() = %v, want %v", gotSelfBinlogCoordinates, tt.wantSelfBinlogCoordinates)
			}
		})
	}
}

func TestGetTableColumns(t *testing.T) {
	type args struct {
		db           *gosql.DB
		databaseName string
		tableName    string
	}
	tests := []struct {
		name    string
		args    args
		want    *umconf.ColumnList
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTableColumns(tt.args.db, tt.args.databaseName, tt.args.tableName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTableColumns() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTableColumns() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplyColumnTypes(t *testing.T) {
	type args struct {
		db           *gosql.Tx
		database     string
		tablename    string
		columnsLists []*umconf.ColumnList
	}
	tests := []struct {
		name    string
		args    args
		want    []*umconf.ColumnList
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ApplyColumnTypes(tt.args.db, tt.args.database, tt.args.tablename, tt.args.columnsLists...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyColumnTypes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.args.columnsLists, tt.want) {
				t.Errorf("ApplyColumnTypes() = %v, want %v", tt.args.columnsLists, tt.want)
			}
		})
	}
}

func TestShowCreateTable(t *testing.T) {
	type args struct {
		db                *gosql.DB
		databaseName      string
		tableName         string
		dropTableIfExists bool
	}
	tests := []struct {
		name                     string
		args                     args
		wantCreateTableStatement string
		wantErr                  bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCreateTableStatement, err := ShowCreateTable(tt.args.db, tt.args.databaseName, tt.args.tableName, tt.args.dropTableIfExists)
			if (err != nil) != tt.wantErr {
				t.Errorf("ShowCreateTable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotCreateTableStatement != tt.wantCreateTableStatement {
				t.Errorf("ShowCreateTable() = %v, want %v", gotCreateTableStatement, tt.wantCreateTableStatement)
			}
		})
	}
}

func Test_parseInterval(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name    string
		args    args
		wantI   gomysql.Interval
		wantErr bool
	}{
		// TODO: Add test cases.
		{"t1", args{"36671-36677"}, gomysql.Interval{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotI, err := parseInterval(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseInterval() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotI, tt.wantI) {
				t.Errorf("parseInterval() = %v, want %v", gotI, tt.wantI)
			}
		})
	}
}

/* NB: SelectGtidExecuted modified. Test case should be changed as well.
func TestSelectGtidExecuted(t *testing.T) {
	uri := "root:rootroot@tcp(192.168.99.100:13309)/?timeout=5s&tls=false&autocommit=true&charset=utf8mb4,utf8,latin1&multiStatements=true"
	db, err := sql.CreateDB(uri)
	if err != nil {
		fmt.Errorf(err.Error())
	}
	type args struct {
		db  *gosql.DB
		sid string
		gno int64
	}
	tests := []struct {
		name        string
		args        args
		wantGtidset string
		wantErr     bool
	}{
		// TODO: Add test cases.
		{"t1", args{db, "96fda9dc-7cbf-11e7-9340-0242ac110002", 1}, "", false}, //96fda9dc-7cbf-11e7-9340-0242ac110002:10116-10120:10126
		{"t2", args{db, "96fda9dc-7cbf-11e7-9340-0242ac110002", 10}, "", false},
		{"t3", args{db, "96fda9dc-7cbf-11e7-9340-0242ac110002", 36678}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotGtidset, err := SelectGtidExecuted(tt.args.db, tt.args.sid, tt.args.gno)
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectGtidExecuted() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotGtidset != tt.wantGtidset {
				t.Errorf("SelectGtidExecuted() = %v, want %v", gotGtidset, tt.wantGtidset)
			}
		})
	}
}
*/

func Test_stringInterval(t *testing.T) {
	type args struct {
		intervals gomysql.IntervalSlice
	}
	tests := []struct {
		name string
		args args
		want string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringInterval(tt.args.intervals); got != tt.want {
				t.Errorf("stringInterval() = %v, want %v", got, tt.want)
			}
		})
	}
}
