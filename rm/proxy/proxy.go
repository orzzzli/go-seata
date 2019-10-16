package proxy

import (
	"errors"
	"goseata/util"
	"strings"
)

type SQLType int

const (
	SQLUnknow         = 0
	SQLUpdate         = 1
	SQLInsert         = 2
	SQLSelect         = 3
	SQLDelete         = 4
	SQLCommit         = 5
	SQLRollback       = 6
	SQLBegin          = 7
	SQLGlobalCommit   = 8
	SQLGlobalRollback = 9
)

type SQLProxy struct {
	Connection string
	Database   string
	Tid        string
	OriginSQL  string
	UpperSQL   string
	SQLType    SQLType
	TableName  string
	WhereStr   string
	ChangeMap  map[string]string
}

func New(tid string, sql string) (*SQLProxy, error) {
	proxy := new(SQLProxy)
	proxy.Tid = tid
	proxy.Connection = "db-user" //todo:改成从配置读取
	proxy.Database = "user"

	reqSql := strings.TrimSpace(sql)   //去掉双端空格
	reqSql = strings.Trim(reqSql, ";") //去掉双端分号

	proxy.OriginSQL = reqSql
	err := proxy.analyseSQLType()
	if err != nil {
		return nil, err
	}
	err = proxy.analyseSQLTable()
	if err != nil {
		return nil, err
	}
	err = proxy.analyseSQLWhere()
	if err != nil {
		return nil, err
	}
	err = proxy.analyseSQLChange()
	if err != nil {
		return nil, err
	}

	return proxy, nil
}

func (p *SQLProxy) analyseSQLType() error {
	reqSqlSlice := strings.Split(p.OriginSQL, " ")      //空格分割
	reqSqlSlice = util.FilterSliceEmptyEle(reqSqlSlice) //去空
	if len(reqSqlSlice) == 0 {
		return errors.New("analyse sql type fail. sql is " + p.OriginSQL)
	}
	method := strings.ToUpper(reqSqlSlice[0])
	p.SQLType = SQLUnknow
	if method == "SELECT" {
		p.SQLType = SQLSelect
	}
	if method == "UPDATE" {
		p.SQLType = SQLUpdate
	}
	if method == "INSERT" {
		p.SQLType = SQLInsert
	}
	if method == "DELETE" {
		p.SQLType = SQLDelete
	}
	if method == "COMMIT" {
		p.SQLType = SQLCommit
	}
	if method == "ROLLBACK" {
		p.SQLType = SQLRollback
	}
	if method == "BEGIN" {
		p.SQLType = SQLBegin
	}
	if method == "GCOMMIT" {
		p.SQLType = SQLGlobalCommit
	}
	if method == "GROLLBACK" {
		p.SQLType = SQLGlobalRollback
	}
	if p.SQLType == SQLUnknow {
		return errors.New("sql type is not support. sql is " + p.OriginSQL)
	}
	return nil
}

func (p *SQLProxy) analyseSQLTable() error {
	// SELECT ... FROM table ....
	if p.SQLType == SQLSelect {
		reqSqlSlice := strings.Split(p.OriginSQL, "FROM") //FROM分割
		if len(reqSqlSlice) < 2 {
			reqSqlSlice = strings.Split(p.OriginSQL, "from") //from分割
			if len(reqSqlSlice) < 2 {
				return errors.New("analyse select sql table fail. sql is " + p.OriginSQL)
			}
		}
		reqSqlSlice = strings.Split(reqSqlSlice[1], " ")    //后半部分空格分割
		reqSqlSlice = util.FilterSliceEmptyEle(reqSqlSlice) //去空
		if len(reqSqlSlice) == 0 {
			return errors.New("analyse select sql table fail. sql is " + p.OriginSQL)
		}
		p.TableName = strings.Trim(reqSqlSlice[0], "`") //去除table的`号
	}
	// UPDATE table SET ....
	if p.SQLType == SQLUpdate {
		reqSqlSlice := strings.Split(p.OriginSQL, " ")      //空格分割
		reqSqlSlice = util.FilterSliceEmptyEle(reqSqlSlice) //去空
		if len(reqSqlSlice) < 2 {
			return errors.New("analyse update sql table fail. sql is " + p.OriginSQL)
		}
		p.TableName = strings.Trim(reqSqlSlice[1], "`") //去除table的`号
	}
	// INSERT INTO table ....
	if p.SQLType == SQLInsert {
		reqSqlSlice := strings.Split(p.OriginSQL, " ")      //空格分割
		reqSqlSlice = util.FilterSliceEmptyEle(reqSqlSlice) //去空
		if len(reqSqlSlice) < 3 {
			return errors.New("analyse insert sql table fail. sql is " + p.OriginSQL)
		}
		p.TableName = strings.Trim(reqSqlSlice[2], "`") //去除table的`号
	}
	// DELETE FROM table ....
	if p.SQLType == SQLDelete {
		reqSqlSlice := strings.Split(p.OriginSQL, " ")      //空格分割
		reqSqlSlice = util.FilterSliceEmptyEle(reqSqlSlice) //去空
		if len(reqSqlSlice) < 3 {
			return errors.New("analyse delete sql table fail. sql is " + p.OriginSQL)
		}
		p.TableName = strings.Trim(reqSqlSlice[2], "`") //去除table的`号
	}
	return nil
}

//只需要select可以直接使用的where语句即可
func (p *SQLProxy) analyseSQLWhere() error {
	p.WhereStr = ""
	//UPDATE table SET ... WHERE ...
	if p.SQLType == SQLUpdate {
		reqSqlSlice := strings.Split(p.OriginSQL, "WHERE") //FROM分割
		if len(reqSqlSlice) < 2 {
			reqSqlSlice = strings.Split(p.OriginSQL, "where") //from分割
			if len(reqSqlSlice) < 2 {
				p.WhereStr = ""
				return nil
			}
		}
		whereStr := strings.TrimSpace(reqSqlSlice[1]) //去掉双端空格
		p.WhereStr = whereStr
	}
	//DELETE FROM table WHERE ...
	if p.SQLType == SQLDelete {
		reqSqlSlice := strings.Split(p.OriginSQL, "WHERE") //FROM分割
		if len(reqSqlSlice) < 2 {
			reqSqlSlice = strings.Split(p.OriginSQL, "where") //from分割
			if len(reqSqlSlice) < 2 {
				p.WhereStr = ""
				return nil
			}
		}
		whereStr := strings.TrimSpace(reqSqlSlice[1]) //去掉双端空格
		p.WhereStr = whereStr
	}
	return nil
}

//解析sql修改的字段，目前只有update
func (p *SQLProxy) analyseSQLChange() error {
	p.ChangeMap = make(map[string]string)
	//UPDATE table SET ... WHERE ...
	if p.SQLType == SQLUpdate {
		reqSqlSlice := strings.Split(p.OriginSQL, "SET") //FROM分割
		if len(reqSqlSlice) < 2 {
			reqSqlSlice = strings.Split(p.OriginSQL, "set") //from分割
			if len(reqSqlSlice) < 2 {
				return errors.New("analyse update sql set fail. sql is " + p.OriginSQL)
			}
		}
		setStr := strings.TrimSpace(reqSqlSlice[1])  //去掉双端空格
		whereIndex := strings.Index(setStr, "where") //是否有where
		if whereIndex == -1 {
			whereIndex = strings.Index(setStr, "WHERE") //是否有WHERE
		}
		//存在where
		if whereIndex != -1 {
			setStr = setStr[0:whereIndex]
			setStr = strings.TrimSpace(setStr) //去掉双端空格
		}
		setStrSlice := strings.Split(setStr, ",") //,分割
		for _, v := range setStrSlice {
			v = strings.TrimSpace(v)
			setItemSlice := strings.Split(v, "=") //=分割
			if len(setItemSlice) < 2 {
				return errors.New("analyse update sql set col fail. sql is " + p.OriginSQL)
			}
			for index, item := range setItemSlice {
				if index%2 == 0 {
					p.ChangeMap[strings.TrimSpace(item)] = strings.TrimSpace(setItemSlice[index+1])
				}
			}
		}
	}
	return nil
}

func (p *SQLProxy) BuildBeforeSelectSQL() string {
	sql := "SELECT id"
	for k, _ := range p.ChangeMap {
		if k == "id" || k == "ID" {
			continue
		}
		sql += "," + k
	}
	sql += " FROM `" + p.TableName + "`"
	if p.WhereStr != "" {
		sql += " WHERE " + p.WhereStr
	}
	sql += " FOR UPDATE;"
	return sql
}

func (p *SQLProxy) BuildAfterSelectSQL() string {
	sql := "SELECT id"
	for k, _ := range p.ChangeMap {
		if k == "id" || k == "ID" {
			continue
		}
		sql += "," + k
	}
	sql += " FROM `" + p.TableName + "`"
	if p.WhereStr != "" {
		sql += " WHERE " + p.WhereStr
	}
	sql += ";"
	return sql
}
