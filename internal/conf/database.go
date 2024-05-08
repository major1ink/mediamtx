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
	UseUpdaterStatus     bool   `json:"useUpdaterStatus"`
	UseSrise             bool   `json:"useSrise"`
	UseProxy             bool   `json:"useProxy"`
	Sql                  Sql    `json:"sql"`
}

type Sql struct {
	InsertPath      string `json:"insertPath"`
	GetPathStream   string `json:"getPathStream"`
	GetCodeMP       string `json:"getCodeMP"`
	UpdateSize      string `json:"updateSize"`
	GetDrives       string `json:"getDrives"`
	UpdateStatus    string `json:"updateStatus"`
	GetData         string `json:"getData"`
	GetDataContract string `json:"getDataContract"`
	GetDataForProxy string `json:"getDataForProxy"`
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
	db.Sql = Sql{
		InsertPath:    "",
		GetPathStream: "",
		UpdateSize:    "",
		GetDrives:     "",
		GetCodeMP:     "",
		UpdateStatus:  "",
	}

}