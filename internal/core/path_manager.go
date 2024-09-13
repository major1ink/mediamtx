package core

import (
	"context"
	"fmt"
	"net/netip"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/bluenviron/mediamtx/internal/auth"
	"github.com/bluenviron/mediamtx/internal/conf"
	"github.com/bluenviron/mediamtx/internal/defs"
	"github.com/bluenviron/mediamtx/internal/externalcmd"
	RMS "github.com/bluenviron/mediamtx/internal/repgrpc"
	"github.com/bluenviron/mediamtx/internal/logger"

	"github.com/bluenviron/mediamtx/internal/stream"

	"github.com/bluenviron/mediamtx/internal/storage"
)

func pathConfCanBeUpdated(oldPathConf *conf.Path, newPathConf *conf.Path) bool {
	clone := oldPathConf.Clone()

	clone.Record = newPathConf.Record

	clone.RPICameraBrightness = newPathConf.RPICameraBrightness
	clone.RPICameraContrast = newPathConf.RPICameraContrast
	clone.RPICameraSaturation = newPathConf.RPICameraSaturation
	clone.RPICameraSharpness = newPathConf.RPICameraSharpness
	clone.RPICameraExposure = newPathConf.RPICameraExposure
	clone.RPICameraFlickerPeriod = newPathConf.RPICameraFlickerPeriod
	clone.RPICameraAWB = newPathConf.RPICameraAWB
	clone.RPICameraAWBGains = newPathConf.RPICameraAWBGains
	clone.RPICameraDenoise = newPathConf.RPICameraDenoise
	clone.RPICameraShutter = newPathConf.RPICameraShutter
	clone.RPICameraMetering = newPathConf.RPICameraMetering
	clone.RPICameraGain = newPathConf.RPICameraGain
	clone.RPICameraEV = newPathConf.RPICameraEV
	clone.RPICameraFPS = newPathConf.RPICameraFPS
	clone.RPICameraIDRPeriod = newPathConf.RPICameraIDRPeriod
	clone.RPICameraBitrate = newPathConf.RPICameraBitrate

	return newPathConf.Equal(clone)
}

type pathManagerHLSServer interface {
	PathReady(defs.Path)
	PathNotReady(defs.Path)
}

type pathManagerParent interface {
	logger.Writer
}

type pathManager struct {
	logLevel          conf.LogLevel
	logDestinations   []logger.Destination
	logFile           string
	logStreams        bool
	logDirStreams     string
	authManager       *auth.Manager
	rtspAddress       string
	readTimeout       conf.StringDuration
	writeTimeout      conf.StringDuration
	writeQueueSize    int
	udpMaxPayloadSize int
	pathConfs         map[string]*conf.Path
	externalCmdPool   *externalcmd.Pool
	parent            pathManagerParent

	ctx         context.Context
	ctxCancel   func()
	wg          sync.WaitGroup
	hlsManager  pathManagerHLSServer
	paths       map[string]*path
	pathsByConf map[string]map[*path]struct{}
	ChConfigSet chan []struct {
		Name   string
		Record bool
	}

	// in

	chReloadConf   chan map[string]*conf.Path
	chSetHLSServer chan pathManagerHLSServer
	chClosePath    chan *path
	chPathReady    chan *path
	chPathNotReady chan *path
	chFindPathConf chan defs.PathFindPathConfReq
	chDescribe     chan defs.PathDescribeReq
	chAddReader    chan defs.PathAddReaderReq
	chAddPublisher chan defs.PathAddPublisherReq
	chAPIPathsList chan pathAPIPathsListReq
	chAPIPathsGet  chan pathAPIPathsGetReq

	clientGRPC      RMS.GrpcClient
	stor      storage.Storage
	switches  conf.Switches
	Publisher MaxPub
	max       int

}

func (pm *pathManager) initialize() {
	ctx, ctxCancel := context.WithCancel(context.Background())

	pm.ctx = ctx
	pm.ctxCancel = ctxCancel
	pm.paths = make(map[string]*path)
	pm.pathsByConf = make(map[string]map[*path]struct{})
	pm.chReloadConf = make(chan map[string]*conf.Path)
	pm.chSetHLSServer = make(chan pathManagerHLSServer)
	pm.chClosePath = make(chan *path)
	pm.chPathReady = make(chan *path)
	pm.chPathNotReady = make(chan *path)
	pm.chFindPathConf = make(chan defs.PathFindPathConfReq)
	pm.chDescribe = make(chan defs.PathDescribeReq)
	pm.chAddReader = make(chan defs.PathAddReaderReq)
	pm.chAddPublisher = make(chan defs.PathAddPublisherReq)
	pm.chAPIPathsList = make(chan pathAPIPathsListReq)
	pm.chAPIPathsGet = make(chan pathAPIPathsGetReq)

	for _, pathConf := range pm.pathConfs {
		if pathConf.Regexp == nil {
			pm.createPath(pathConf, pathConf.Name, nil)
		}
	}

	pm.Log(logger.Debug, "path manager created")

	pm.wg.Add(1)
	go pm.run()
}

type bdTable struct {
	Id             int
	Login          string
	Pass           string
	Ip_address_out netip.Prefix
	Cam_path       string
	Code_mp        string
	State_public   int
	Status_public  int
	Contract       string
	Record         bool
}

type prohys struct {
	Ip_address_out string
	Code_mp        string
}

func getTypeInt(item interface{}) int {

	t := reflect.TypeOf(item)

	if t.Kind() == reflect.Int8 {
		return int(item.(int8))
	}

	if t.Kind() == reflect.Int16 {
		return int(item.(int16))
	}

	if t.Kind() == reflect.Int32 {
		return int(item.(int32))
	}

	return int(item.(int64))
}

func getTypeBool(item interface{}) bool {

	s := getTypeInt(item)
	return s == 1
}

func (pm *pathManager) close() {
	pm.Log(logger.Debug, "path manager is shutting down")
	pm.ctxCancel()
	pm.wg.Wait()
}

// Log implements logger.Writer.
func (pm *pathManager) Log(level logger.Level, format string, args ...interface{}) {
	pm.parent.Log(level, format, args...)
}

func transformation (i int16)bool{
	if i == 0 {
		return false
	}
	if i== 1 {
		return true 
	}
	return false
}
func (pm *pathManager) checkStatus() {
	outer:
		for {
			select {
			case <-time.After(time.Duration(pm.switches.TimeStatus) * time.Second):
				
				if len(pm.paths) > 0 {
					switch{
						case pm.clientGRPC.Use:
							var relodSt []struct {
							Name string
							Record bool
							}
							for name := range pm.paths {
								pm.Log(logger.Debug, "sending a request to receive a Status Record via stream: %s", name)
								r,err:=pm.clientGRPC.Select(name,"StatusRecord")
								if err != nil {
									pm.Log(logger.Error, "%v", err)
									continue
								}
								pm.Log(logger.Debug, "StatusRecord = %v",r.StatusRecord)
								if transformation(int16(r.StatusRecord)) != pm.paths[name].conf.Record {
							pm.paths[name].Log(logger.Debug, "[record] status_record = %v",r.StatusRecord)
							relodSt = append(relodSt, struct{Name string; Record bool}{Name: name, Record: transformation(int16(r.StatusRecord))})
								}

							}
							if len(relodSt) > 0 {
								pm.ChConfigSet <- relodSt
							}

						case pm.stor.Use:

												query := pm.stor.Sql.GetStatus_records + "("
					for name := range pm.paths {
						query = query + "'" + name + "'" + ","
					}
					query = query[:len(query)-1] + ")"
					pm.Log(logger.Debug, fmt.Sprintf("SQL query sent:%s", query))
					data, err := pm.stor.Req.SelectData(query)
					if err != nil {
						pm.Log(logger.Error, "%v", err)
						continue outer
					}
					pm.Log(logger.Debug, "The result of executing the sql query: %v", data)
					var relodSt []struct {
						Name string
						Record bool
					}
					for _, resul := range data {
						if transformation(resul[0].(int16)) != pm.paths[resul[1].(string)].conf.Record {
							pm.paths[resul[1].(string)].Log(logger.Debug, "[record] status_record = %v",resul[0].(int16))
							relodSt = append(relodSt, struct{Name string; Record bool}{Name: resul[1].(string), Record: transformation(resul[0].(int16))})
						}
					}

					if len(relodSt) > 0 {
					pm.ChConfigSet <- relodSt
					}
					}

				}
			case <-pm.ctx.Done():
				break outer
			}
		}
	}
func (pm *pathManager) run() {
	defer pm.wg.Done()

	if pm.switches.UsePathStream && (pm.stor.Use || pm.clientGRPC.Use ){
		go pm.checkStatus()
	}
outer:

	for {
		select {
		case newPaths := <-pm.chReloadConf:
			pm.doReloadConf(newPaths)

		case m := <-pm.chSetHLSServer:
			pm.doSetHLSServer(m)

		case pa := <-pm.chClosePath:
			pm.doClosePath(pa)

		case pa := <-pm.chPathReady:
			pm.doPathReady(pa)

		case pa := <-pm.chPathNotReady:
			pm.doPathNotReady(pa)

		case req := <-pm.chFindPathConf:
			pm.doFindPathConf(req)

		case req := <-pm.chDescribe:
			pm.doDescribe(req)

		case req := <-pm.chAddReader:
			pm.doAddReader(req)

		case req := <-pm.chAddPublisher:
			pm.doAddPublisher(req)

		case req := <-pm.chAPIPathsList:
			pm.doAPIPathsList(req)

		case req := <-pm.chAPIPathsGet:
			pm.doAPIPathsGet(req)

		case <-pm.ctx.Done():
			break outer
		}
	}

	pm.ctxCancel()
}

func (pm *pathManager) doReloadConf(newPaths map[string]*conf.Path) {
	for confName, pathConf := range pm.pathConfs {
		if newPath, ok := newPaths[confName]; ok {
			// configuration has changed
			if !newPath.Equal(pathConf) {
				if pathConfCanBeUpdated(pathConf, newPath) { // paths associated with the configuration can be updated
					for pa := range pm.pathsByConf[confName] {
						go pa.reloadConf(newPath)
					}
				} else { // paths associated with the configuration must be recreated
					for pa := range pm.pathsByConf[confName] {
						pm.removePath(pa)
						pa.close()
						pa.wait() // avoid conflicts between sources
					}
				}
			}
		} else {
			// configuration has been deleted, remove associated paths
			for pa := range pm.pathsByConf[confName] {
				pm.removePath(pa)
				pa.close()
				pa.wait() // avoid conflicts between sources
			}
		}
	}

	pm.pathConfs = newPaths

	// add new paths
	for pathConfName, pathConf := range pm.pathConfs {
		if _, ok := pm.paths[pathConfName]; !ok && pathConf.Regexp == nil {
			pm.createPath(pathConf, pathConfName, nil)
		}
	}
}

func (pm *pathManager) doSetHLSServer(m pathManagerHLSServer) {
	pm.hlsManager = m
}

func (pm *pathManager) doClosePath(pa *path) {
	if pmpa, ok := pm.paths[pa.name]; !ok || pmpa != pa {
		return
	}
	pm.removePath(pa)
}

func (pm *pathManager) doPathReady(pa *path) {
	if pm.hlsManager != nil {
		pm.hlsManager.PathReady(pa)
	}
}

func (pm *pathManager) doPathNotReady(pa *path) {
	if pm.hlsManager != nil {
		pm.hlsManager.PathNotReady(pa)
	}
}

func (pm *pathManager) doFindPathConf(req defs.PathFindPathConfReq) {
	pathConf, _, err := conf.FindPathConf(pm.pathConfs, req.AccessRequest.Name)
	if err != nil {
		req.Res <- defs.PathFindPathConfRes{Err: err}
		return
	}

	err = pm.authManager.Authenticate(req.AccessRequest.ToAuthRequest())
	if err != nil {
		req.Res <- defs.PathFindPathConfRes{Err: err}
		return
	}

	req.Res <- defs.PathFindPathConfRes{Conf: pathConf}
}

func (pm *pathManager) doDescribe(req defs.PathDescribeReq) {
	pathConf, pathMatches, err := conf.FindPathConf(pm.pathConfs, req.AccessRequest.Name)
	if err != nil {
		req.Res <- defs.PathDescribeRes{Err: err}
		return
	}
	err = pm.authManager.Authenticate(req.AccessRequest.ToAuthRequest())
	if err != nil {
		req.Res <- defs.PathDescribeRes{Err: err}
		return
	}

	// create path if it doesn't exist
	if _, ok := pm.paths[req.AccessRequest.Name]; !ok {
		pm.createPath(pathConf, req.AccessRequest.Name, pathMatches)
	}

	req.Res <- defs.PathDescribeRes{Path: pm.paths[req.AccessRequest.Name]}
}

func (pm *pathManager) doAddReader(req defs.PathAddReaderReq) {
	pathConf, pathMatches, err := conf.FindPathConf(pm.pathConfs, req.AccessRequest.Name)
	if err != nil {
		req.Res <- defs.PathAddReaderRes{Err: err}
		return
	}
	if !req.AccessRequest.SkipAuth {
		err = pm.authManager.Authenticate(req.AccessRequest.ToAuthRequest())
		if err != nil {
			req.Res <- defs.PathAddReaderRes{Err: err}
			return
		}
	}

	// create path if it doesn't exist
	if _, ok := pm.paths[req.AccessRequest.Name]; !ok {
		pm.createPath(pathConf, req.AccessRequest.Name, pathMatches)
	}

	req.Res <- defs.PathAddReaderRes{Path: pm.paths[req.AccessRequest.Name]}
}

func (pm *pathManager) doAddPublisher(req defs.PathAddPublisherReq) {
	pathConf, pathMatches, err := conf.FindPathConf(pm.pathConfs, req.AccessRequest.Name)
	if err != nil {
		req.Res <- defs.PathAddPublisherRes{Err: err}
		return
	}

	if !req.AccessRequest.SkipAuth {
		err = pm.authManager.Authenticate(req.AccessRequest.ToAuthRequest())
		if err != nil {
			req.Res <- defs.PathAddPublisherRes{Err: err}
			return
		}
	}

	// create path if it doesn't exist
	if _, ok := pm.paths[req.AccessRequest.Name]; !ok {
		pm.createPath(pathConf, req.AccessRequest.Name, pathMatches)
	}
	req.Res <- defs.PathAddPublisherRes{Path: pm.paths[req.AccessRequest.Name]}
}

func (pm *pathManager) doAPIPathsList(req pathAPIPathsListReq) {
	paths := make(map[string]*path)

	for name, pa := range pm.paths {
		paths[name] = pa
	}

	req.res <- pathAPIPathsListRes{paths: paths}
}

func (pm *pathManager) doAPIPathsGet(req pathAPIPathsGetReq) {
	path, ok := pm.paths[req.name]
	if !ok {
		req.res <- pathAPIPathsGetRes{err: conf.ErrPathNotFound}
		return
	}

	req.res <- pathAPIPathsGetRes{path: path}
}

func (pm *pathManager) createPath(
	pathConf *conf.Path,
	name string,
	matches []string,
) {

	pa := &path{
		parentCtx:         pm.ctx,
		logLevel:          pm.logLevel,
		rtspAddress:       pm.rtspAddress,
		readTimeout:       pm.readTimeout,
		writeTimeout:      pm.writeTimeout,
		writeQueueSize:    pm.writeQueueSize,
		udpMaxPayloadSize: pm.udpMaxPayloadSize,
		conf:              pathConf,
		name:              name,
		matches:           matches,
		wg:                &pm.wg,
		externalCmdPool:   pm.externalCmdPool,
		parent:            pm,
		clientGRPC:        pm.clientGRPC,
		stor:              pm.stor,
		switches: 		   pm.switches,
		publisher:         &pm.Publisher,
		logStreams:        pm.logStreams,
		ChConfigSet:       pm.ChConfigSet,
	}
	if pm.logStreams {
		
		logg, err := logger.NewLoggerStream(logger.Level(pm.logLevel), pm.logDestinations, pm.logFile, name, pm.logDirStreams)
		if err != nil {
			pm.Log(logger.Error, "%s", err)
		}
		pa.loggerPath = logg
	}
	pa.initialize(pm.stor, &pm.Publisher)

	pm.paths[name] = pa

	if _, ok := pm.pathsByConf[pathConf.Name]; !ok {
		pm.pathsByConf[pathConf.Name] = make(map[*path]struct{})
	}
	pm.pathsByConf[pathConf.Name][pa] = struct{}{}
}

func (pm *pathManager) removePath(pa *path) {
	delete(pm.pathsByConf[pa.conf.Name], pa)
	if len(pm.pathsByConf[pa.conf.Name]) == 0 {
		delete(pm.pathsByConf, pa.conf.Name)
	}
	if pm.switches.UseUpdaterStatus && pm.stor.Use{
		query := fmt.Sprintf(pm.stor.Sql.UpdateStatus, 0, pa.Name())
		pa.Log(logger.Debug, "SQL status %s", query)
		err := pm.stor.Req.ExecQuery(query)
		if err != nil {
			pa.Log(logger.Error, "%s", err)
		}
		pa.Log(logger.Debug, "The request was successfully completed")
	}
	if pm.switches.UseUpdaterStatus && pm.stor.Use{
		query := fmt.Sprintf(pm.stor.Sql.UpdateStatus, 0, pa.Name())
		pa.Log(logger.Debug, "SQL status %s", query)
		err := pm.stor.Req.ExecQuery(query)
		if err != nil {
			pa.Log(logger.Error, "%s", err)
		}
		pa.Log(logger.Debug, "The request was successfully completed")
	}
	delete(pm.paths, pa.name)
}

// ReloadPathConfs is called by core.
func (pm *pathManager) ReloadPathConfs(pathConfs map[string]*conf.Path) {
	select {
	case pm.chReloadConf <- pathConfs:
	case <-pm.ctx.Done():
	}
}

// pathReady is called by path.
func (pm *pathManager) pathReady(pa *path) {
	select {
	case pm.chPathReady <- pa:
		if pm.switches.UseUpdaterStatus && pm.stor.Use{
			query := fmt.Sprintf(pm.stor.Sql.UpdateStatus, 1, pa.Name())
			pa.Log(logger.Debug, "SQL status %s", query)
			err := pm.stor.Req.ExecQuery(query)
			if err != nil {
				pa.Log(logger.Error, "%s", err)
			}
			pa.Log(logger.Debug, "The request was successfully completed")
		}
	case <-pm.ctx.Done():
	case <-pa.ctx.Done(): // in case pathManager is blocked by path.wait()
	}
}

// pathNotReady is called by path.
func (pm *pathManager) pathNotReady(pa *path) {
	select {
	case pm.chPathNotReady <- pa:
		if pm.switches.UseUpdaterStatus && pm.stor.Use{
			query := fmt.Sprintf(pm.stor.Sql.UpdateStatus, 0, pa.Name())
			pa.Log(logger.Debug, "SQL status %s", query)
			err := pm.stor.Req.ExecQuery(query)
			if err != nil {
				pa.Log(logger.Error, "%s", err)
			}
			pa.Log(logger.Debug, "The request was successfully completed")
		}
	case <-pm.ctx.Done():
	case <-pa.ctx.Done(): // in case pathManager is blocked by path.wait()
	}
}

// closePath is called by path.
func (pm *pathManager) closePath(pa *path) {
	select {
	case pm.chClosePath <- pa:
		if pm.switches.UseUpdaterStatus && pm.stor.Use{
			query := fmt.Sprintf(pm.stor.Sql.UpdateStatus, 0, pa.Name())
			pa.Log(logger.Debug, "SQL status %s", query)
			err := pm.stor.Req.ExecQuery(query)
			if err != nil {
				pa.Log(logger.Error, "%s", err)
			}
			pa.Log(logger.Debug, "The request was successfully completed")
		}
	case <-pm.ctx.Done():
	case <-pa.ctx.Done(): // in case pathManager is blocked by path.wait()
	}
}

// GetConfForPath is called by a reader or publisher.
func (pm *pathManager) FindPathConf(req defs.PathFindPathConfReq) (*conf.Path, error) {
	req.Res = make(chan defs.PathFindPathConfRes)
	select {
	case pm.chFindPathConf <- req:
		res := <-req.Res
		return res.Conf, res.Err

	case <-pm.ctx.Done():
		return nil, fmt.Errorf("terminated")
	}
}

// Describe is called by a reader or publisher.
func (pm *pathManager) Describe(req defs.PathDescribeReq) defs.PathDescribeRes {
	req.Res = make(chan defs.PathDescribeRes)
	select {
	case pm.chDescribe <- req:
		res1 := <-req.Res
		if res1.Err != nil {
			return res1
		}

		res2 := res1.Path.(*path).describe(req)
		if res2.Err != nil {
			return res2
		}

		res2.Path = res1.Path
		return res2

	case <-pm.ctx.Done():
		return defs.PathDescribeRes{Err: fmt.Errorf("terminated")}
	}
}

// AddPublisher is called by a publisher.
func (pm *pathManager) AddPublisher(req defs.PathAddPublisherReq) (defs.Path, error) {
	req.Res = make(chan defs.PathAddPublisherRes)
	select {
	case pm.chAddPublisher <- req:
		res := <-req.Res
		
		if res.Err != nil {
			return nil, res.Err
		}

		return res.Path.(*path).addPublisher(req)

	case <-pm.ctx.Done():
		return nil, fmt.Errorf("terminated")
	}
}

// AddReader is called by a reader.
func (pm *pathManager) AddReader(req defs.PathAddReaderReq) (defs.Path, *stream.Stream, error) {
	req.Res = make(chan defs.PathAddReaderRes)
	select {
	case pm.chAddReader <- req:
		res := <-req.Res
		if res.Err != nil {
			return nil, nil, res.Err
		}

		return res.Path.(*path).addReader(req)

	case <-pm.ctx.Done():
		return nil, nil, fmt.Errorf("terminated")
	}
}

// setHLSServer is called by hlsManager.
func (pm *pathManager) setHLSServer(s pathManagerHLSServer) {
	select {
	case pm.chSetHLSServer <- s:
	case <-pm.ctx.Done():
	}
}

// APIPathsList is called by api.
func (pm *pathManager) APIPathsList() (*defs.APIPathList, error) {
	req := pathAPIPathsListReq{
		res: make(chan pathAPIPathsListRes),
	}

	select {
	case pm.chAPIPathsList <- req:
		res := <-req.res

		res.data = &defs.APIPathList{
			Items: []*defs.APIPath{},
		}

		for _, pa := range res.paths {
			item, err := pa.APIPathsGet(pathAPIPathsGetReq{})
			if err == nil {
				res.data.Items = append(res.data.Items, item)
			}
		}

		sort.Slice(res.data.Items, func(i, j int) bool {
			return res.data.Items[i].Name < res.data.Items[j].Name
		})

		return res.data, nil

	case <-pm.ctx.Done():
		return nil, fmt.Errorf("terminated")
	}
}

// APIPathsGet is called by api.
func (pm *pathManager) APIPathsGet(name string) (*defs.APIPath, error) {
	req := pathAPIPathsGetReq{
		name: name,
		res:  make(chan pathAPIPathsGetRes),
	}

	select {
	case pm.chAPIPathsGet <- req:
		res := <-req.res
		if res.err != nil {
			return nil, res.err
		}

		data, err := res.path.APIPathsGet(req)
		return data, err

	case <-pm.ctx.Done():
		return nil, fmt.Errorf("terminated")
	}
}
