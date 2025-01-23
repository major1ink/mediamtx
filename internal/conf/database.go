package conf

type Database struct {
	Use                  bool   `json:"use"`
	DbAddress            string `json:"dbAddress"`
	DbPort               int    `json:"dbPort"`
	DbName               string `json:"dbName"`
	DbUser               string `json:"dbUser"`
	DbPassword           string `json:"dbPassword"`
	MaxConnections       int    `json:"maxConnections"`
	Sql                  Sql    `json:"sql"`
}

type Sql struct {
	InsertPath        string `json:"insertPath"`
	InsertPathStream string `json:"insertPathStream"`
	GetPathStream     string `json:"getPathStream"`
	GetCodeMP         string `json:"getCodeMP"`
	GetDrives         string `json:"getDrives"`
	UpdateStatus      string `json:"updateStatus"`
	GetDataForProxy   string `json:"getDataForProxy"`
	GetStatus_records string `json:"getStatus_records"`
}

func (db *Database) setDefaults() {

	db.Use = false
	db.DbAddress = "127.0.0.1"
	db.DbPort = 5432
	db.DbName = "postgres"
	db.DbUser = "postgres"
	db.DbPassword = ""
	db.MaxConnections = 1
	db.Sql = Sql{
		InsertPath:    "",
		InsertPathStream: "",
		GetPathStream: "",
		GetDrives:     "",
		GetCodeMP:     "",
		UpdateStatus:  "",
		GetDataForProxy: "",
		GetStatus_records: "",
	}

}
