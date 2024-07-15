package conf

type Database struct {
	Use                  bool   `json:"use"`
	DbAddress            string `json:"dbAddress"`
	DbPort               int    `json:"dbPort"`
	DbName               string `json:"dbName"`
	DbUser               string `json:"dbUser"`
	DbPassword           string `json:"dbPassword"`
	MaxConnections       int    `json:"maxConnections"`
	DbDrives             bool   `json:"dbDrives"`
	DbUseCodeMP_Contract bool   `json:"dbUseCodeMP_Contract"`
	DbUseContract        bool   `json:"dbUseContract"`
	UseDbPathStream      bool   `json:"useDbPathStream"`
	TimeStatus           int    `json:"timeStatus"`
	UseUpdaterStatus     bool   `json:"useUpdaterStatus"`
	UseSrise             bool   `json:"useSrise"`
	UseProxy             bool   `json:"useProxy"`
	Login                string `json:"login"`
	Pass                 string `json:"pass"`
	FileSQLErr           string `json:"fileSQLErr"`
	Sql                  Sql    `json:"sql"`
}

type Sql struct {
	InsertPath        string `json:"insertPath"`
	InsertPathStream string `json:"insertPathStream"`
	GetPathStream     string `json:"getPathStream"`
	GetCodeMP         string `json:"getCodeMP"`
	GetDrives         string `json:"getDrives"`
	UpdateStatus      string `json:"updateStatus"`
	GetData           string `json:"getData"`
	GetDataContract   string `json:"getDataContract"`
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
	db.MaxConnections = 0
	db.DbDrives = false
	db.DbUseCodeMP_Contract = false
	db.UseDbPathStream = false
	db.UseUpdaterStatus = false
	db.UseSrise = false
	db.UseProxy = false
	db.Login = ""
	db.Pass = ""
	db.FileSQLErr = "./"
	db.Sql = Sql{
		InsertPath:    "",
		GetPathStream: "",
		GetDrives:     "",
		GetCodeMP:     "",
		UpdateStatus:  "",
	}

}
