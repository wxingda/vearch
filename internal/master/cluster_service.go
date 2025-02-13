// Copyright 2019 The Vearch Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package master

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/cubefs/cubefs/depends/tiglabs/raft/proto"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"github.com/vearch/vearch/v3/internal/client"
	"github.com/vearch/vearch/v3/internal/config"
	"github.com/vearch/vearch/v3/internal/entity"
	"github.com/vearch/vearch/v3/internal/pkg/errutil"
	"github.com/vearch/vearch/v3/internal/pkg/log"
	"github.com/vearch/vearch/v3/internal/pkg/metrics/mserver"
	"github.com/vearch/vearch/v3/internal/pkg/number"
	"github.com/vearch/vearch/v3/internal/pkg/vjson"
	"github.com/vearch/vearch/v3/internal/proto/vearchpb"
	"github.com/vearch/vearch/v3/internal/ps/engine/mapping"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

// masterService is used for master administrator purpose. It should not used by router or partition server program
type masterService struct {
	*client.Client
}

func newMasterService(client *client.Client) (*masterService, error) {
	return &masterService{client}, nil
}

// registerServerService find nodeId partitions
func (ms *masterService) registerServerService(ctx context.Context, ip string, nodeID entity.NodeID) (*entity.Server, error) {
	server := &entity.Server{Ip: ip}

	spaces, err := ms.Master().QuerySpacesByKey(ctx, entity.PrefixSpace)
	if err != nil {
		return nil, err
	}

	for _, s := range spaces {
		for _, p := range s.Partitions {
			for _, id := range p.Replicas {
				if nodeID == id {
					server.PartitionIds = append(server.PartitionIds, p.Id)
					server.Spaces = append(server.Spaces, s)
					break
				}
			}
		}
	}

	return server, nil
}

// registerPartitionService partition/[id]:[body]
func (ms *masterService) registerPartitionService(ctx context.Context, partition *entity.Partition) error {
	log.Info("register partition:[%d] ", partition.Id)
	marshal, err := vjson.Marshal(partition)
	if err != nil {
		return err
	}
	return ms.Master().Put(ctx, entity.PartitionKey(partition.Id), marshal)
}

// createDBService three keys "db/id/[dbId]:[dbName]" ,"db/name/[dbName]:[dbId]" ,"db/body/[dbId]:[dbBody]"
func (ms *masterService) createDBService(ctx context.Context, db *entity.DB) (err error) {
	if ms.Master().Client().Master().Config().Global.LimitedDBNum {
		_, bytesArr, err := ms.Master().PrefixScan(ctx, entity.PrefixDataBaseBody)
		if err != nil {
			return err
		}
		if len(bytesArr) >= 1 {
			return fmt.Errorf("db num is limited to one and already have one db exists")
		}
	}

	//validate name has in db is in return err
	if err = db.Validate(); err != nil {
		return err
	}

	if err = ms.validatePS(ctx, db.Ps); err != nil {
		return err
	}

	//generate a new db id
	db.Id, err = ms.Master().NewIDGenerate(ctx, entity.DBIdSequence, 1, 5*time.Second)
	if err != nil {
		return err
	}

	// it will lock cluster to create db
	mutex := ms.Master().NewLock(ctx, entity.LockDBKey(db.Name), time.Second*300)
	if err = mutex.Lock(); err != nil {
		return err
	}
	defer func() {
		if err := mutex.Unlock(); err != nil {
			log.Error("unlock db err:[%s]", err.Error())
		}
	}()
	err = ms.Master().STM(context.Background(), func(stm concurrency.STM) error {
		idKey, nameKey, bodyKey := ms.Master().DBKeys(db.Id, db.Name)

		if stm.Get(nameKey) != "" {
			return fmt.Errorf("dbname %s is exists", db.Name)
		}

		if stm.Get(idKey) != "" {
			return fmt.Errorf("dbID %d is exists", db.Id)
		}
		value, err := vjson.Marshal(db)
		if err != nil {
			return err
		}
		stm.Put(nameKey, cast.ToString(db.Id))
		stm.Put(idKey, db.Name)
		stm.Put(bodyKey, string(value))
		return nil
	})
	return err
}

func (ms *masterService) validatePS(ctx context.Context, psList []string) error {
	if len(psList) < 1 {
		return nil
	}
	servers, err := ms.Master().QueryServers(ctx)
	if err != nil {
		return err
	}
	for _, ps := range psList {
		flag := false
		for _, server := range servers {
			if server.Ip == ps {
				flag = true
				if !client.IsLive(server.RpcAddr()) {
					return fmt.Errorf("server:[%s] can not connection", ps)
				}
				break
			}
		}
		if !flag {
			return fmt.Errorf("server:[%s] not found in cluster", ps)
		}
	}
	return nil
}

func (ms *masterService) deleteDBService(ctx context.Context, dbstr string) (err error) {
	db, err := ms.queryDBService(ctx, dbstr)
	if err != nil {
		return err
	}
	// it will lock cluster to delete db
	mutex := ms.Master().NewLock(ctx, entity.LockDBKey(dbstr), time.Second*300)
	if err = mutex.Lock(); err != nil {
		return err
	}
	defer func() {
		if err := mutex.Unlock(); err != nil {
			log.Error("unlock db err:[%s]", err.Error())
		}
	}()

	spaces, err := ms.Master().QuerySpaces(ctx, db.Id)
	if err != nil {
		return err
	}

	if len(spaces) > 0 {
		return vearchpb.NewError(vearchpb.ErrorEnum_DB_NOT_EMPTY, nil)
	}

	err = ms.Master().STM(context.Background(),
		func(stm concurrency.STM) error {
			idKey, nameKey, bodyKey := ms.Master().DBKeys(db.Id, db.Name)
			stm.Del(idKey)
			stm.Del(nameKey)
			stm.Del(bodyKey)
			return nil
		})

	if err != nil {
		return err
	}

	return nil
}

func (ms *masterService) updateDBIpList(ctx context.Context, dbModify *entity.DBModify) (db *entity.DB, err error) {
	// process panic
	defer errutil.CatchError(&err)
	var id int64
	db = &entity.DB{}
	if number.IsNum(dbModify.DbName) {
		id = cast.ToInt64(dbModify.DbName)
	} else if id, err = ms.Master().QueryDBName2Id(ctx, dbModify.DbName); err != nil {
		return nil, err
	}
	bs, err := ms.Master().Get(ctx, entity.DBKeyBody(id))
	errutil.ThrowError(err)
	if bs == nil {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_DB_NOT_EXIST, nil)
	}
	err = vjson.Unmarshal(bs, db)
	errutil.ThrowError(err)
	if dbModify.Method == proto.ConfRemoveNode {
		ps := make([]string, 0, len(db.Ps))
		for _, ip := range db.Ps {
			if ip != dbModify.IPAddr {
				ps = append(ps, ip)
			}
		}
		db.Ps = ps
	} else if dbModify.Method == proto.ConfAddNode {
		exist := false
		for _, ip := range db.Ps {
			if ip == dbModify.IPAddr {
				exist = true
				break
			}
		}
		if !exist {
			db.Ps = append(db.Ps, dbModify.IPAddr)
		}
	} else {
		err = fmt.Errorf("method support addIP:[%d] removeIP:[%d] not support:[%d]",
			proto.ConfAddNode, proto.ConfRemoveNode, dbModify.Method)
		errutil.ThrowError(err)
	}
	log.Debug("db info is %v", db)
	_, _, bodyKey := ms.Master().DBKeys(db.Id, db.Name)
	value, err := vjson.Marshal(db)
	errutil.ThrowError(err)
	err = ms.Client.Master().Put(ctx, bodyKey, value)
	return db, err
}

func (ms *masterService) queryDBService(ctx context.Context, dbstr string) (db *entity.DB, err error) {
	var id int64
	db = &entity.DB{}
	if number.IsNum(dbstr) {
		id = cast.ToInt64(dbstr)
	} else if id, err = ms.Master().QueryDBName2Id(ctx, dbstr); err != nil {
		return nil, err
	}

	bs, err := ms.Master().Get(ctx, entity.DBKeyBody(id))

	if err != nil {
		return nil, err
	}

	if bs == nil {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_DB_NOT_EXIST, nil)
	}

	if err := vjson.Unmarshal(bs, db); err != nil {
		return nil, err
	} else {
		return db, nil
	}
}

// server/[serverAddr]:[serverBody]
// spaceKeys "space/[dbId]/[spaceId]:[spaceBody]"
func (ms *masterService) createSpaceService(ctx context.Context, dbName string, space *entity.Space) (err error) {
	if space.DBId, err = ms.Master().QueryDBName2Id(ctx, dbName); err != nil {
		log.Error("find DbId according to DbName:%v failed, error: %v", dbName, err)
		return err
	}

	// to validate schema
	_, err = mapping.SchemaMap(space.Fields)
	if err != nil {
		log.Error("master service createSpaceService error: %v", err)
		return err
	}

	// it will lock cluster to create space
	mutex := ms.Master().NewLock(ctx, entity.LockSpaceKey(dbName, spaceName), time.Second*300)
	if err = mutex.Lock(); err != nil {
		return err
	}
	defer func() {
		if err := mutex.Unlock(); err != nil {
			log.Error("unlock space err:[%s]", err.Error())
		}
	}()

	// spaces is existed
	if _, err := ms.Master().QuerySpaceByName(ctx, space.DBId, space.Name); err != nil {
		vErr := vearchpb.NewError(vearchpb.ErrorEnum_INTERNAL_ERROR, err)
		if vErr.GetError().Code != vearchpb.ErrorEnum_SPACE_NOT_EXIST {
			return vErr
		}
	} else {
		return vearchpb.NewError(vearchpb.ErrorEnum_SPACE_EXIST, nil)
	}

	log.Info("create space, db: %s, spaceName: %s, space :[%s]", dbName, space.Name, vjson.ToJsonString(space))

	// find all servers for create space
	servers, err := ms.Master().QueryServers(ctx)
	if err != nil {
		return err
	}

	// generate space id
	spaceID, err := ms.Master().NewIDGenerate(ctx, entity.SpaceIdSequence, 1, 5*time.Second)
	if err != nil {
		return err
	}
	space.Id = spaceID

	spaceProperties, err := entity.UnmarshalPropertyJSON(space.Fields)
	if err != nil {
		return err
	}

	space.SpaceProperties = spaceProperties
	for _, f := range spaceProperties {
		if f.FieldType == vearchpb.FieldType_VECTOR && f.Index != nil {
			space.Index = f.Index
		}
	}

	if space.PartitionRule != nil {
		err := space.PartitionRule.Validate(space, true)
		if err != nil {
			return err
		}
		width := math.MaxUint32 / (space.PartitionNum * space.PartitionRule.Partitions)
		for i := 0; i < space.PartitionNum*space.PartitionRule.Partitions; i++ {
			partitionID, err := ms.Master().NewIDGenerate(ctx, entity.PartitionIdSequence, 1, 5*time.Second)

			if err != nil {
				return err
			}

			space.Partitions = append(space.Partitions, &entity.Partition{
				Id:      entity.PartitionID(partitionID),
				Name:    space.PartitionRule.Ranges[i/space.PartitionNum].Name,
				SpaceId: space.Id,
				DBId:    space.DBId,
				Slot:    entity.SlotID(i * width),
			})
		}
	} else {
		width := math.MaxUint32 / space.PartitionNum
		for i := 0; i < space.PartitionNum; i++ {
			partitionID, err := ms.Master().NewIDGenerate(ctx, entity.PartitionIdSequence, 1, 5*time.Second)

			if err != nil {
				return err
			}

			space.Partitions = append(space.Partitions, &entity.Partition{
				Id:      entity.PartitionID(partitionID),
				SpaceId: space.Id,
				DBId:    space.DBId,
				Slot:    entity.SlotID(i * width),
			})
		}
	}

	serverPartitions, err := ms.filterAndSortServer(ctx, space, servers)
	if err != nil {
		return err
	}

	if int(space.ReplicaNum) > len(serverPartitions) {
		return fmt.Errorf("not enough partition servers, need %d replicas but only have %d",
			int(space.ReplicaNum), len(serverPartitions))
	}

	bFlase := false
	space.Enabled = &bFlase
	defer func() {
		if !(*space.Enabled) { // remove the space if it is still not enabled
			if e := ms.Master().Delete(context.Background(), entity.SpaceKey(space.DBId, space.Id)); e != nil {
				log.Error("to delete space err: %s", e.Error())
			}
		}
	}()

	marshal, err := vjson.Marshal(space)
	if err != nil {
		return err
	}
	if space.Index == nil {
		return vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("space vector field index should not be empty"))
	}
	err = ms.Master().Create(ctx, entity.SpaceKey(space.DBId, space.Id), marshal)
	if err != nil {
		return err
	}

	// pick servers for space
	var pAddrs [][]string
	for i := 0; i < len(space.Partitions); i++ {
		if addrs, err := ms.selectServersForPartition(servers, serverPartitions, space.ReplicaNum, space.Partitions[i]); err != nil {
			return err
		} else {
			pAddrs = append(pAddrs, addrs)
		}
	}

	var errChain = make(chan error, 1)
	// send create space request to partition server
	for i := 0; i < len(space.Partitions); i++ {
		go func(addrs []string, partition *entity.Partition) {
			defer func() {
				if r := recover(); r != nil {
					err := vearchpb.NewError(vearchpb.ErrorEnum_INTERNAL_ERROR, fmt.Errorf("create partition err: %v", r))
					errChain <- err
					log.Error(err.Error())
				}
			}()
			for _, addr := range addrs {
				if err := client.CreatePartition(addr, space, partition.Id); err != nil {
					err := vearchpb.NewError(vearchpb.ErrorEnum_INTERNAL_ERROR, fmt.Errorf("create partition err: %s", err.Error()))
					errChain <- err
					log.Error(err.Error())
				}
			}
		}(pAddrs[i], space.Partitions[i])
	}

	// check all partition is ok
	for i := 0; i < len(space.Partitions); i++ {
		v := 0
		for {
			v++
			select {
			case err := <-errChain:
				return err
			case <-ctx.Done():
				return vearchpb.NewError(vearchpb.ErrorEnum_INTERNAL_ERROR, fmt.Errorf("create space has error"))
			default:
			}

			partition, err := ms.Master().QueryPartition(ctx, space.Partitions[i].Id)
			if v%5 == 0 {
				log.Debug("check the partition:%d status", space.Partitions[i].Id)
			}
			if err != nil && vearchpb.NewError(vearchpb.ErrorEnum_INTERNAL_ERROR, err).GetError().Code != vearchpb.ErrorEnum_PARTITION_NOT_EXIST {
				return err
			}
			if partition != nil {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
	}

	bTrue := true
	space.Enabled = &bTrue

	// update version
	err = ms.updateSpace(ctx, space)
	if err != nil {
		bFalse := false
		space.Enabled = &bFalse
		return err
	}

	return nil
}

// selectServersForPartition selects servers for a partition based on the given criteria.
// It ensures that the servers with the fewest replicas are chosen and applies anti-affinity by zone if configured.
//
// Parameters:
// - servers: A slice of pointers to Server entities representing available servers.
// - serverPartitions: A map where the key is the server index and the value is the number of partitions on that server.
// - replicaNum: The number of replicas needed for the partition.
// - partition: A pointer to the Partition entity that needs to be assigned servers.
//
// Returns:
// - A slice of strings containing the addresses of the selected servers.
// - An error if the required number of servers could not be selected.
//
// The function considers the anti-affinity strategy configured in the master service to avoid placing replicas in the same zone.
func (ms *masterService) selectServersForPartition(servers []*entity.Server, serverPartitions map[int]int, replicaNum uint8, partition *entity.Partition) ([]string, error) {
	address := make([]string, 0, replicaNum)
	originReplicaNum := replicaNum
	partition.Replicas = make([]entity.NodeID, 0, replicaNum)

	kvList := make([]struct {
		index  int
		length int
	}, len(serverPartitions))

	i := 0
	for k, v := range serverPartitions {
		kvList[i] = struct {
			index  int
			length int
		}{index: k, length: v}
		i++
	}

	sort.Slice(kvList, func(i, j int) bool {
		return kvList[i].length < kvList[j].length
	})

	zoneCount := make(map[string]int)

	antiAffinity := ms.Master().Client().Master().Config().PS.ReplicaAntiAffinityStrategy
	// find the servers with the fewest replicas and apply anti-affinity by zone
	for _, kv := range kvList {
		addr := servers[kv.index].RpcAddr()
		ID := servers[kv.index].ID
		var zone string

		switch antiAffinity {
		case 1:
			zone = servers[kv.index].HostIp
		case 2:
			zone = servers[kv.index].HostRack
		case 3:
			zone = servers[kv.index].HostZone
		default:
			zone = ""
		}

		if !client.IsLive(addr) {
			serverPartitions[kv.index] = kv.length
			continue
		}

		if zone != "" && zoneCount[zone] > 0 {
			continue
		}

		serverPartitions[kv.index]++
		if zone != "" {
			zoneCount[zone]++
		}
		address = append(address, addr)
		partition.Replicas = append(partition.Replicas, ID)

		replicaNum--
		if replicaNum <= 0 {
			break
		}
	}

	if replicaNum > 0 {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_MASTER_PS_NOT_ENOUGH_SELECT, fmt.Errorf("need %d partition servers but only got %d", originReplicaNum, len(address)))
	}

	return address, nil
}

func (ms *masterService) filterAndSortServer(ctx context.Context, space *entity.Space, servers []*entity.Server) (map[int]int, error) {
	db, err := ms.queryDBService(ctx, cast.ToString(space.DBId))
	if err != nil {
		return nil, err
	}

	var psMap map[string]bool
	if len(db.Ps) > 0 {
		psMap = make(map[string]bool)
		for _, ps := range db.Ps {
			psMap[ps] = true
		}
	}

	serverPartitions := make(map[int]int)

	spaces, err := ms.Master().QuerySpacesByKey(ctx, entity.PrefixSpace)
	if err != nil {
		return nil, err
	}

	serverIndex := make(map[entity.NodeID]int)

	if psMap == nil { // If psMap is nil, only use public servers
		for i, s := range servers {
			// Only use servers with the same resource name
			if s.ResourceName != space.ResourceName {
				continue
			}
			if !s.Private {
				serverPartitions[i] = 0
				serverIndex[s.ID] = i
			}
		}
	} else { // If psMap is not nil, only use defined servers
		for i, s := range servers {
			// Only use servers with the same resource name
			if s.ResourceName != space.ResourceName {
				psMap[s.Ip] = false
				continue
			}
			if psMap[s.Ip] {
				serverPartitions[i] = 0
				serverIndex[s.ID] = i
			}
		}
	}

	for _, space := range spaces {
		for _, partition := range space.Partitions {
			for _, nodeID := range partition.Replicas {
				if index, ok := serverIndex[nodeID]; ok {
					serverPartitions[index] = serverPartitions[index] + 1
				}
			}
		}
	}

	return serverPartitions, nil
}

func (ms *masterService) deleteSpaceService(ctx context.Context, dbName string, spaceName string) error {
	dbId, err := ms.Master().QueryDBName2Id(ctx, dbName)
	if err != nil {
		return err
	}

	space, err := ms.Master().QuerySpaceByName(ctx, dbId, spaceName)
	if err != nil {
		return err
	}
	if space == nil { // nil if it not exists
		return nil
	}

	mutex := ms.Master().NewLock(ctx, entity.LockSpaceKey(dbName, spaceName), time.Second*60)
	if err = mutex.Lock(); err != nil {
		return err
	}
	defer func() {
		if err := mutex.Unlock(); err != nil {
			log.Error("unlock space err:[%s]", err.Error())
		}
	}()
	// delete key
	err = ms.Master().Delete(ctx, entity.SpaceKey(dbId, space.Id))
	if err != nil {
		return err
	}

	// delete parition and partitionKey
	for _, p := range space.Partitions {
		for _, replica := range p.Replicas {
			if server, err := ms.Master().QueryServer(ctx, replica); err != nil {
				log.Error("query partition:[%d] for replica:[%s] has err:[%s]", p.Id, replica, err.Error())
			} else {
				if err := client.DeletePartition(server.RpcAddr(), p.Id); err != nil {
					log.Error("delete partition:[%d] for server:[%s] has err:[%s]", p.Id, server.RpcAddr(), err.Error())
				}
			}
		}
		err = ms.Master().Delete(ctx, entity.PartitionKey(p.Id))
		if err != nil {
			return err
		}
	}

	// delete alias
	if aliases, err := ms.queryAllAlias(ctx); err != nil {
		return err
	} else {
		for _, alias := range aliases {
			if alias.DbName == dbName && alias.SpaceName == spaceName {
				if err := ms.deleteAliasService(ctx, alias.Name); err != nil {
					return err
				}
			}
		}
	}
	err = ms.Master().Delete(ctx, entity.SpaceConfigKey(dbId, space.Id))
	if err != nil {
		return err
	}

	return nil
}

func (ms *masterService) queryDBs(ctx context.Context) ([]*entity.DB, error) {
	_, bytesArr, err := ms.Master().PrefixScan(ctx, entity.PrefixDataBaseBody)
	if err != nil {
		return nil, err
	}
	dbs := make([]*entity.DB, 0, len(bytesArr))
	for _, bs := range bytesArr {
		db := &entity.DB{}
		if err := vjson.Unmarshal(bs, db); err != nil {
			log.Error("decode db err: %s,and the bs is:%s", err.Error(), string(bs))
			continue
		}
		dbs = append(dbs, db)
	}

	return dbs, err
}

func (ms *masterService) describeSpaceService(ctx context.Context, space *entity.Space, spaceInfo *entity.SpaceInfo, detail_info bool) (int, error) {
	spaceStatus := 0
	color := []string{"green", "yellow", "red"}
	spaceInfo.Errors = make([]string, 0)

	// check partition num in meta data
	if space.PartitionRule != nil {
		if len(space.Partitions) != int(space.PartitionNum*space.PartitionRule.Partitions) {
			msg := fmt.Sprintf("space: [%s] partitions length:[%d] not equal to partition num:[%d] * PartitionRule.Partitions: [%d]", space.Name, len(space.Partitions), space.PartitionNum, space.PartitionRule.Partitions)
			spaceInfo.Errors = append(spaceInfo.Errors, msg)
			log.Error(msg)
			spaceStatus = 2
		}
	} else {
		if len(space.Partitions) != int(space.PartitionNum) {
			msg := fmt.Sprintf("space: [%s] partitions length:[%d] not equal to partition num:[%d]", space.Name, len(space.Partitions), space.PartitionNum)
			spaceInfo.Errors = append(spaceInfo.Errors, msg)
			log.Error(msg)
			spaceStatus = 2
		}
	}

	for _, spacePartition := range space.Partitions {
		p, err := ms.Master().QueryPartition(ctx, spacePartition.Id)
		pStatus := 0

		if err != nil {
			msg := fmt.Sprintf("partition:[%d] in space: [%s] not found in meta data", spacePartition.Id, space.Name)
			spaceInfo.Errors = append(spaceInfo.Errors, msg)
			log.Error(msg)
			pStatus = 2
			if pStatus > spaceStatus {
				spaceStatus = pStatus
			}
			continue
		}

		nodeID := p.LeaderID
		if nodeID == 0 {
			log.Error("partition:[%d] in space: [%s] leaderID is 0", spacePartition.Id, space.Name)
			if len(p.Replicas) > 0 {
				nodeID = p.Replicas[0]
			}
		}

		server, err := ms.Master().QueryServer(ctx, nodeID)
		if err != nil {
			msg := fmt.Sprintf("space: [%s] partition:[%d], server:[%d] not found", space.Name, spacePartition.Id, nodeID)
			spaceInfo.Errors = append(spaceInfo.Errors, msg)
			log.Error(msg)
			pStatus = 2
			if pStatus > spaceStatus {
				spaceStatus = pStatus
			}
			continue
		}

		partitionInfo, err := client.PartitionInfo(server.RpcAddr(), p.Id, detail_info)
		if err != nil {
			msg := fmt.Sprintf("query space:[%s] server:[%d] partition:[%d] info err :[%s]", space.Name, nodeID, spacePartition.Id, err.Error())
			spaceInfo.Errors = append(spaceInfo.Errors, msg)
			log.Error(msg)
			partitionInfo = &entity.PartitionInfo{}
			pStatus = 2
		} else {
			if len(partitionInfo.Unreachable) > 0 {
				pStatus = 1
			}
		}

		replicasStatus := make(map[entity.NodeID]string)
		for nodeID, status := range p.ReStatusMap {
			if status == entity.ReplicasOK {
				replicasStatus[nodeID] = "ReplicasOK"
			} else {
				replicasStatus[nodeID] = "ReplicasNotReady"
			}
		}

		if partitionInfo.RaftStatus != nil {
			if partitionInfo.RaftStatus.Leader == 0 {
				msg := fmt.Sprintf("partition:[%d] in space:[%s] has no leader", spacePartition.Id, space.Name)
				spaceInfo.Errors = append(spaceInfo.Errors, msg)
				log.Error(msg)
				pStatus = 2
			} else {
				if len(partitionInfo.RaftStatus.Replicas) != int(space.ReplicaNum) {
					msg := fmt.Sprintf("partition:[%d] in space:[%s] replicas: [%d] is not equal to replicaNum: [%d]", spacePartition.Id, space.Name, len(partitionInfo.RaftStatus.Replicas), space.ReplicaNum)
					spaceInfo.Errors = append(spaceInfo.Errors, msg)
					log.Error(msg)
					pStatus = 2
				} else {
					replicaStateProbeNum := 0
					leaderId := 0
					for nodeID, replica := range partitionInfo.RaftStatus.Replicas {
						// TODO FIXME: when leader changed, the unreachableNodeIDnre state may still be ReplicaStateProbe
						isNodeUnreachable := false
						for _, unreachableNodeID := range partitionInfo.Unreachable {
							if nodeID == unreachableNodeID {
								isNodeUnreachable = true
								break
							}
						}
						if isNodeUnreachable {
							continue
						}
						if replica.State == entity.ReplicaStateProbe {
							replicaStateProbeNum += 1
							leaderId = int(nodeID)
						}
					}
					if replicaStateProbeNum != 1 {
						msg := fmt.Sprintf("partition:[%d] in space:[%s] have [%d] leader", spacePartition.Id, space.Name, replicaStateProbeNum)
						spaceInfo.Errors = append(spaceInfo.Errors, msg)
						log.Error(msg)
						pStatus = 2
					}
					if leaderId != int(partitionInfo.RaftStatus.Leader) {
						msg := fmt.Sprintf("partition:[%d] in space:[%s] leader: [%d] is not equal to raft leader: [%d]", spacePartition.Id, space.Name, leaderId, partitionInfo.RaftStatus.Leader)
						spaceInfo.Errors = append(spaceInfo.Errors, msg)
						log.Error(msg)
						pStatus = 2
					}
				}
			}
		}

		//this must from space.Partitions
		partitionInfo.PartitionID = spacePartition.Id
		partitionInfo.Name = spacePartition.Name
		partitionInfo.Color = color[pStatus]
		partitionInfo.ReplicaNum = len(p.Replicas)
		partitionInfo.Ip = server.Ip
		partitionInfo.NodeID = server.ID
		partitionInfo.RepStatus = replicasStatus

		spaceInfo.Partitions = append(spaceInfo.Partitions, partitionInfo)

		if pStatus > spaceStatus {
			spaceStatus = pStatus
		}
	}

	docNum := uint64(0)
	for _, p := range spaceInfo.Partitions {
		docNum += cast.ToUint64(p.DocNum)
	}
	spaceInfo.Status = color[spaceStatus]
	spaceInfo.DocNum = docNum
	return spaceStatus, nil
}

// createAliasService keys "/alias/alias_name:alias"
func (ms *masterService) createAliasService(ctx context.Context, alias *entity.Alias) (err error) {
	//validate name
	if err = alias.Validate(); err != nil {
		return err
	}
	mutex := ms.Master().NewLock(ctx, entity.LockAliasKey(alias.Name), time.Second*30)
	if err = mutex.Lock(); err != nil {
		return err
	}
	defer func() {
		if err := mutex.Unlock(); err != nil {
			log.Error("unlock lock for create alias err %s", err)
		}
	}()
	err = ms.Master().STM(context.Background(), func(stm concurrency.STM) error {
		aliasKey := entity.AliasKey(alias.Name)

		value := stm.Get(aliasKey)
		if value != "" {
			return vearchpb.NewError(vearchpb.ErrorEnum_ALIAS_EXIST, nil)
		}
		marshal, err := vjson.Marshal(alias)
		if err != nil {
			return err
		}
		stm.Put(aliasKey, string(marshal))
		return nil
	})
	return err
}

func (ms *masterService) deleteAliasService(ctx context.Context, alias_name string) (err error) {
	alias, err := ms.queryAliasService(ctx, alias_name)
	if err != nil {
		return err
	}
	//it will lock cluster
	mutex := ms.Master().NewLock(ctx, entity.LockAliasKey(alias_name), time.Second*30)
	if err = mutex.Lock(); err != nil {
		return err
	}
	defer func() {
		if err := mutex.Unlock(); err != nil {
			log.Error("unlock lock for delete alias err %s", err)
		}
	}()

	err = ms.Master().STM(context.Background(),
		func(stm concurrency.STM) error {
			stm.Del(entity.AliasKey(alias.Name))
			return nil
		})

	if err != nil {
		return err
	}

	return nil
}

func (ms *masterService) updateAliasService(ctx context.Context, alias *entity.Alias) (err error) {
	bs, err := ms.Master().Get(ctx, entity.AliasKey(alias.Name))
	if err != nil {
		return err
	}
	if bs == nil {
		return vearchpb.NewError(vearchpb.ErrorEnum_ALIAS_NOT_EXIST, nil)
	}

	//it will lock cluster
	mutex := ms.Master().NewLock(ctx, entity.LockAliasKey(alias.Name), time.Second*30)
	if err = mutex.Lock(); err != nil {
		return err
	}
	defer func() {
		if err := mutex.Unlock(); err != nil {
			log.Error("unlock lock for update alias err %s", err)
		}
	}()
	err = ms.Master().STM(context.Background(), func(stm concurrency.STM) error {
		marshal, err := vjson.Marshal(alias)
		if err != nil {
			return err
		}
		stm.Put(entity.AliasKey(alias.Name), string(marshal))
		return nil
	})
	return nil
}

func (ms *masterService) queryAllAlias(ctx context.Context) ([]*entity.Alias, error) {
	_, values, err := ms.Master().PrefixScan(ctx, entity.PrefixAlias)
	if err != nil {
		return nil, err
	}
	allAlias := make([]*entity.Alias, 0, len(values))
	for _, value := range values {
		if value == nil {
			return nil, vearchpb.NewError(vearchpb.ErrorEnum_ALIAS_NOT_EXIST, nil)
		}
		alias := &entity.Alias{}
		err = vjson.Unmarshal(value, alias)
		if err != nil {
			return nil, vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("get alias:%s value:%s, err:%s", alias.Name, string(value), err.Error()))
		}
		allAlias = append(allAlias, alias)
	}

	return allAlias, err
}

func (ms *masterService) queryAliasService(ctx context.Context, alias_name string) (alias *entity.Alias, err error) {
	alias = &entity.Alias{Name: alias_name}

	bs, err := ms.Master().Get(ctx, entity.AliasKey(alias_name))

	if err != nil {
		return nil, err
	}

	if bs == nil {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_ALIAS_NOT_EXIST, nil)
	}

	err = vjson.Unmarshal(bs, alias)
	if err != nil {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("get alias:%s value:%s, err:%s", alias.Name, string(bs), err.Error()))
	}
	return alias, nil
}

// createUserService keys "/user/user_name:user"
func (ms *masterService) createUserService(ctx context.Context, user *entity.User, check_root bool) (err error) {
	//validate name
	if err = user.Validate(check_root); err != nil {
		return err
	}

	if user.RoleName != nil {
		if _, err := ms.queryRoleService(ctx, *user.RoleName); err != nil {
			return err
		}
	} else {
		return vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("role name is empty"))
	}
	if user.Password == nil {
		return vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("password is empty"))
	}

	mutex := ms.Master().NewLock(ctx, entity.LockUserKey(user.Name), time.Second*30)
	if err = mutex.Lock(); err != nil {
		return err
	}
	defer func() {
		if err := mutex.Unlock(); err != nil {
			log.Error("unlock lock for create user err %s", err)
		}
	}()
	err = ms.Master().STM(context.Background(), func(stm concurrency.STM) error {
		userKey := entity.UserKey(user.Name)

		value := stm.Get(userKey)
		if value != "" {
			return vearchpb.NewError(vearchpb.ErrorEnum_USER_EXIST, nil)
		}
		marshal, err := vjson.Marshal(user)
		if err != nil {
			return err
		}
		stm.Put(userKey, string(marshal))
		return nil
	})
	return err
}

func (ms *masterService) deleteUserService(ctx context.Context, user_name string) (err error) {
	if strings.EqualFold(user_name, entity.RootName) {
		return vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("can't delete root user"))
	}
	user, err := ms.queryUserService(ctx, user_name, false)
	if user == nil {
		return err
	}
	//it will lock cluster
	mutex := ms.Master().NewLock(ctx, entity.LockUserKey(user_name), time.Second*30)
	if err = mutex.Lock(); err != nil {
		return err
	}
	defer func() {
		if err := mutex.Unlock(); err != nil {
			log.Error("unlock lock for delete user err %s", err)
		}
	}()

	err = ms.Master().STM(context.Background(),
		func(stm concurrency.STM) error {
			stm.Del(entity.UserKey(user.Name))
			return nil
		})

	if err != nil {
		return err
	}

	return nil
}

func (ms *masterService) updateUserService(ctx context.Context, user *entity.User, auth_user string) (err error) {
	bs, err := ms.Master().Get(ctx, entity.UserKey(user.Name))
	if err != nil {
		return err
	}
	if bs == nil {
		return vearchpb.NewError(vearchpb.ErrorEnum_USER_NOT_EXIST, nil)
	}

	old_user := &entity.User{}
	err = vjson.Unmarshal(bs, old_user)
	if err != nil {
		return vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("updata user:%s err:%s", user.Name, err.Error()))
	}

	if user.RoleName != nil {
		if user.Password != nil || user.OldPassword != nil {
			return vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("don't update role or password at same time"))
		}
		if _, err := ms.queryRoleService(ctx, *user.RoleName); err != nil {
			return err
		}
		if old_user.Password != nil {
			user.Password = old_user.Password
		}
	} else {
		if auth_user == entity.RootName && user.Name != entity.RootName {
			if user.Password == nil {
				return vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("empty password"))
			}
			if old_user.Password != nil && *user.Password == *old_user.Password {
				return vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("password is same with old password"))
			}
		} else {
			if user.Password == nil || user.OldPassword == nil {
				return vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("empty password or old password"))
			}
			if old_user.Password != nil && *user.Password == *old_user.Password {
				return vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("password is same with old password"))
			}
			if old_user.Password != nil && *user.OldPassword != *old_user.Password {
				return vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("old password is invalid"))
			}
		}

		if old_user.RoleName != nil {
			user.RoleName = old_user.RoleName
		}
	}

	//it will lock cluster
	mutex := ms.Master().NewLock(ctx, entity.LockUserKey(user.Name), time.Second*30)
	if err = mutex.Lock(); err != nil {
		return err
	}
	defer func() {
		if err := mutex.Unlock(); err != nil {
			log.Error("unlock lock for update alias err %s", err)
		}
	}()
	err = ms.Master().STM(context.Background(), func(stm concurrency.STM) error {
		marshal, err := vjson.Marshal(user)
		if err != nil {
			return err
		}
		stm.Put(entity.UserKey(user.Name), string(marshal))
		return nil
	})
	return nil
}

func (ms *masterService) queryAllUser(ctx context.Context) ([]*entity.UserRole, error) {
	_, values, err := ms.Master().PrefixScan(ctx, entity.PrefixUser)
	if err != nil {
		return nil, err
	}
	users := make([]*entity.UserRole, 0, len(values))
	for _, value := range values {
		if value == nil {
			return nil, vearchpb.NewError(vearchpb.ErrorEnum_USER_NOT_EXIST, nil)
		}
		user := &entity.User{}
		err = vjson.Unmarshal(value, user)
		if err != nil {
			return nil, vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("get user:%s, err:%s", user.Name, err.Error()))
		}
		userRole := &entity.UserRole{Name: user.Name}
		if role, err := ms.queryRoleService(ctx, *user.RoleName); err != nil {
			return nil, vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("get user:%s role:%s, err:%s", user.Name, *user.RoleName, err.Error()))
		} else {
			userRole.Role = *role
		}

		users = append(users, userRole)
	}

	return users, err
}

func (ms *masterService) queryUserService(ctx context.Context, user_name string, check_role bool) (userRole *entity.UserRole, err error) {
	user := &entity.User{Name: user_name}

	bs, err := ms.Master().Get(ctx, entity.UserKey(user_name))

	if err != nil {
		return nil, err
	}

	if bs == nil {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_USER_NOT_EXIST, nil)
	}

	err = vjson.Unmarshal(bs, user)
	if err != nil {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("get user:%s, err:%s", user.Name, err.Error()))
	}

	userRole = &entity.UserRole{Name: user.Name}
	if check_role {
		if role, err := ms.queryRoleService(ctx, *user.RoleName); err != nil {
			return nil, vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("get user:%s role:%s, err:%s", user.Name, *user.RoleName, err.Error()))
		} else {
			userRole.Role = *role
		}
	}

	return userRole, nil
}

func (ms *masterService) queryUserWithPasswordService(ctx context.Context, user_name string, check_role bool) (userRole *entity.UserRole, err error) {
	user := &entity.User{Name: user_name}

	bs, err := ms.Master().Get(ctx, entity.UserKey(user_name))

	if err != nil {
		return nil, err
	}

	if bs == nil {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_USER_NOT_EXIST, nil)
	}

	err = vjson.Unmarshal(bs, user)
	if err != nil {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("get user:%s, err:%s", user.Name, err.Error()))
	}

	userRole = &entity.UserRole{Name: user.Name, Password: user.Password}
	if check_role {
		if role, err := ms.queryRoleService(ctx, *user.RoleName); err != nil {
			return nil, vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("get user:%s role:%s, err:%s", user.Name, *user.RoleName, err.Error()))
		} else {
			userRole.Role = *role
		}
	}

	return userRole, nil
}

// createRoleService keys "/role/role_name:role"
func (ms *masterService) createRoleService(ctx context.Context, role *entity.Role) (err error) {
	//validate
	if err = role.Validate(); err != nil {
		return err
	}
	mutex := ms.Master().NewLock(ctx, entity.LockRoleKey(role.Name), time.Second*30)
	if err = mutex.Lock(); err != nil {
		return err
	}
	defer func() {
		if err := mutex.Unlock(); err != nil {
			log.Error("unlock lock for create role err %s", err)
		}
	}()
	err = ms.Master().STM(context.Background(), func(stm concurrency.STM) error {
		roleKey := entity.RoleKey(role.Name)

		value := stm.Get(roleKey)
		if value != "" {
			return vearchpb.NewError(vearchpb.ErrorEnum_ROLE_EXIST, nil)
		}
		marshal, err := vjson.Marshal(role)
		if err != nil {
			return err
		}
		stm.Put(roleKey, string(marshal))
		return nil
	})
	return err
}

func (ms *masterService) deleteRoleService(ctx context.Context, role_name string) (err error) {
	role, err := ms.queryRoleService(ctx, role_name)
	if err != nil {
		return err
	}
	//it will lock cluster
	mutex := ms.Master().NewLock(ctx, entity.LockRoleKey(role_name), time.Second*30)
	if err = mutex.Lock(); err != nil {
		return err
	}
	defer func() {
		if err := mutex.Unlock(); err != nil {
			log.Error("unlock lock for delete role err %s", err)
		}
	}()

	err = ms.Master().STM(context.Background(),
		func(stm concurrency.STM) error {
			stm.Del(entity.RoleKey(role.Name))
			return nil
		})

	if err != nil {
		return err
	}

	return nil
}

func (ms *masterService) changeRolePrivilegeService(ctx context.Context, role *entity.Role) (new_role *entity.Role, err error) {
	//validate
	if err = role.Validate(); err != nil {
		return nil, err
	}

	bs, err := ms.Master().Get(ctx, entity.RoleKey(role.Name))
	if err != nil {
		return nil, err
	}
	if bs == nil {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_ROLE_NOT_EXIST, nil)
	}
	old_role := &entity.Role{}
	err = vjson.Unmarshal(bs, old_role)
	if err != nil {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("get role privilege:%s err:%s", old_role.Name, err.Error()))
	}

	//it will lock cluster
	mutex := ms.Master().NewLock(ctx, entity.LockRoleKey(role.Name), time.Second*30)
	if err = mutex.Lock(); err != nil {
		return nil, err
	}
	defer func() {
		if err := mutex.Unlock(); err != nil {
			log.Error("unlock lock for update role privilege err %s", err)
		}
	}()
	err = ms.Master().STM(context.Background(), func(stm concurrency.STM) error {
		if len(old_role.Privileges) == 0 {
			old_role.Privileges = make(map[entity.Resource]entity.Privilege)
		}
		for resource, privilege := range role.Privileges {
			if role.Operator == entity.Grant {
				old_role.Privileges[resource] = privilege
			}
			if role.Operator == entity.Revoke {
				delete(old_role.Privileges, resource)
			}
		}
		marshal, err := vjson.Marshal(old_role)
		if err != nil {
			return err
		}
		stm.Put(entity.RoleKey(old_role.Name), string(marshal))
		return nil
	})
	return old_role, nil
}

func (ms *masterService) queryAllRole(ctx context.Context) ([]*entity.Role, error) {
	_, values, err := ms.Master().PrefixScan(ctx, entity.PrefixRole)
	if err != nil {
		return nil, err
	}
	roles := make([]*entity.Role, 0, len(values))
	for _, value := range values {
		if value == nil {
			return nil, vearchpb.NewError(vearchpb.ErrorEnum_ROLE_NOT_EXIST, nil)
		}
		role := &entity.Role{}
		err = vjson.Unmarshal(value, role)
		if err != nil {
			return nil, vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("get role:%s, err:%s", role.Name, err.Error()))
		}
		roles = append(roles, role)
	}

	return roles, err
}

func (ms *masterService) queryRoleService(ctx context.Context, role_name string) (role *entity.Role, err error) {
	role = &entity.Role{Name: role_name}

	if value, exists := entity.RoleMap[role_name]; exists {
		role = &value
		return role, nil
	}

	bs, err := ms.Master().Get(ctx, entity.RoleKey(role_name))

	if err != nil {
		return nil, err
	}

	if bs == nil {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_ROLE_NOT_EXIST, nil)
	}

	err = vjson.Unmarshal(bs, role)
	if err != nil {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("get role:%s, err:%s", role.Name, err.Error()))
	}

	return role, nil
}

func (ms *masterService) GetEngineCfg(ctx context.Context, dbName, spaceName string) (cfg *entity.EngineConfig, err error) {
	defer errutil.CatchError(&err)
	// get space info
	dbId, err := ms.Master().QueryDBName2Id(ctx, dbName)
	if err != nil {
		errutil.ThrowError(err)
	}

	space, err := ms.Master().QuerySpaceByName(ctx, dbId, spaceName)
	if err != nil {
		log.Error("query space %s/%s err: %s", dbName, spaceName, err.Error())
		return nil, err
	}
	cfg, err = ms.getEngineConfig(ctx, space)

	if err == nil {
		return cfg, nil
	}

	// invoke all space nodeID
	if space != nil && space.Partitions != nil {
		for _, partition := range space.Partitions {
			// get all replicas nodeID
			if partition.Replicas != nil {
				for _, nodeID := range partition.Replicas {
					server, err := ms.Master().QueryServer(ctx, nodeID)
					errutil.ThrowError(err)
					// send rpc query
					log.Debug("invoke nodeID [%+v],address [%+v]", partition.Id, server.RpcAddr())
					cfg, err = client.GetEngineCfg(server.RpcAddr(), partition.Id)
					errutil.ThrowError(err)
					if err == nil {
						return cfg, nil
					}
				}
			}
		}
	}
	return nil, fmt.Errorf("cannot get engine config")
}

func (ms *masterService) ModifyEngineCfg(ctx context.Context, dbName, spaceName string, cfg *entity.EngineConfig) (err error) {
	defer errutil.CatchError(&err)
	// get space info
	dbId, err := ms.Master().QueryDBName2Id(ctx, dbName)
	if err != nil {
		errutil.ThrowError(err)
	}

	space, err := ms.Master().QuerySpaceByName(ctx, dbId, spaceName)
	if err != nil {
		errutil.ThrowError(err)
	}
	// invoke all space nodeID
	if space != nil && space.Partitions != nil {
		for _, partition := range space.Partitions {
			// get all replicas nodeID
			if partition.Replicas != nil {
				for _, nodeID := range partition.Replicas {
					server, err := ms.Master().QueryServer(ctx, nodeID)
					errutil.ThrowError(err)
					// send rpc query
					log.Debug("invoke nodeID [%+v],address [%+v]", partition.Id, server.RpcAddr())
					err = client.UpdateEngineCfg(server.RpcAddr(), cfg, partition.Id)
					errutil.ThrowError(err)
				}
			}
		}
	}

	err = ms.updateEngineConfig(ctx, space, cfg)
	if err != nil {
		log.Error("update engine config err: %s", err.Error())
		return err
	}
	return nil
}

func (ms *masterService) updateSpaceService(ctx context.Context, dbName, spaceName string, temp *entity.Space) (*entity.Space, error) {
	// it will lock cluster to create space
	mutex := ms.Master().NewLock(ctx, entity.LockSpaceKey(dbName, spaceName), time.Second*300)
	if err := mutex.Lock(); err != nil {
		return nil, err
	}
	defer func() {
		if err := mutex.Unlock(); err != nil {
			log.Error("failed to unlock space,the Error is:%v ", err)
		}
	}()

	dbId, err := ms.Master().QueryDBName2Id(ctx, dbName)
	if err != nil {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("failed to find database id according database name:%v,the Error is:%v ", dbName, err))
	}

	space, err := ms.Master().QuerySpaceByName(ctx, dbId, spaceName)
	if err != nil {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("failed to find space according space name:%v,the Error is:%v ", spaceName, err))
	}

	if space == nil {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("can not found space by name : %s", spaceName))
	}

	buff := bytes.Buffer{}
	if temp.DBId != 0 && temp.DBId != space.DBId {
		buff.WriteString("db_id not same ")
	}

	if temp.PartitionNum != 0 && temp.PartitionNum != space.PartitionNum {
		buff.WriteString("partition_num can not change ")
	}
	if temp.ReplicaNum != 0 && temp.ReplicaNum != space.ReplicaNum {
		buff.WriteString("replica_num can not change ")
	}
	if buff.String() != "" {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf(buff.String()))
	}

	if temp.Name != "" {
		space.Name = temp.Name
	}

	if temp.Enabled != nil {
		space.Enabled = temp.Enabled
	}

	if err := space.Validate(); err != nil {
		return nil, err
	}

	space.Version++
	space.Partitions = temp.Partitions

	if temp.Fields != nil && len(temp.Fields) > 0 {
		// parse old space
		oldFieldMap, err := mapping.SchemaMap(space.Fields)
		if err != nil {
			return nil, err
		}

		// parse new space
		newFieldMap, err := mapping.SchemaMap(temp.Fields)
		if err != nil {
			return nil, err
		}

		for k, v := range oldFieldMap {
			if fm, ok := newFieldMap[k]; ok {
				if !mapping.Equals(v, fm) {
					return nil, vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("not equals by field:[%s] old[%v] new[%v]", k, v, fm))
				}
				delete(newFieldMap, k)
			}
		}

		if len(newFieldMap) > 0 {
			log.Info("change schema for space: %s, change fields: %d, value is: [%s]", space.Name, len(newFieldMap), string(temp.Fields))

			schema, err := mapping.MergeSchema(space.Fields, temp.Fields)
			if err != nil {
				return nil, err
			}

			space.Fields = schema
		}
	}

	// notify all partitions
	for _, p := range space.Partitions {
		partition, err := ms.Master().QueryPartition(ctx, p.Id)
		if err != nil {
			return nil, err
		}

		server, err := ms.Master().QueryServer(ctx, partition.LeaderID)
		if err != nil {
			return nil, err
		}

		if !client.IsLive(server.RpcAddr()) {
			return nil, vearchpb.NewError(vearchpb.ErrorEnum_PARTITION_IS_CLOSED, fmt.Errorf("partition %s is shutdown", server.RpcAddr()))
		}
	}

	for _, p := range space.Partitions {
		partition, err := ms.Master().QueryPartition(ctx, p.Id)
		if err != nil {
			return nil, err
		}

		server, err := ms.Master().QueryServer(ctx, partition.LeaderID)
		if err != nil {
			return nil, err
		}

		log.Debug("update partition server is [%+v], space is [%+v], pid is [%+v]",
			server, space, p.Id)

		if err := client.UpdatePartition(server.RpcAddr(), space, p.Id); err != nil {
			log.Error("UpdatePartition err is [%v]", err)
			return nil, err
		}
	}

	log.Debug("update space is [%+v]", space)
	space.Version--
	if err := ms.updateSpace(ctx, space); err != nil {
		return nil, err
	}

	return space, nil
}

func (ms *masterService) updateSpace(ctx context.Context, space *entity.Space) error {
	space.Version++
	if space.PartitionRule == nil {
		space.PartitionNum = len(space.Partitions)
	}
	marshal, err := vjson.Marshal(space)
	if err != nil {
		return err
	}
	if err = ms.Master().Update(ctx, entity.SpaceKey(space.DBId, space.Id), marshal); err != nil {
		return err
	}

	return nil
}

func (ms *masterService) getEngineConfig(ctx context.Context, space *entity.Space) (cfg *entity.EngineConfig, err error) {
	defer errutil.CatchError(&err)
	marshal, err := ms.Master().Get(ctx, entity.SpaceConfigKey(space.DBId, space.Id))
	if err != nil {
		return nil, err
	}

	cfg = &entity.EngineConfig{}
	if err = vjson.Unmarshal(marshal, cfg); err != nil {
		return nil, err
	}

	return cfg, err
}

func (ms *masterService) updateEngineConfig(ctx context.Context, space *entity.Space, cfg *entity.EngineConfig) error {
	old_cfg, err := ms.getEngineConfig(ctx, space)
	if err != nil {
		log.Error("get engine config err: %s", err.Error())
	}

	new_cfg := cfg
	if old_cfg != nil {
		new_cfg = old_cfg
		if cfg.EngineCacheSize != nil {
			new_cfg.EngineCacheSize = cfg.EngineCacheSize
		}
		if cfg.LongSearchTime != nil {
			new_cfg.LongSearchTime = cfg.LongSearchTime
		}
		if cfg.Path != nil {
			new_cfg.Path = cfg.Path
		}
	}
	marshal, err := vjson.Marshal(new_cfg)
	if err != nil {
		return err
	}
	if err = ms.Master().Update(ctx, entity.SpaceConfigKey(space.DBId, space.Id), marshal); err != nil {
		return err
	}

	return nil
}

func (ms *masterService) BackupSpace(ctx context.Context, dbName, spaceName string, backup *entity.BackupSpace) (err error) {
	clusterName := config.Conf().Global.Name

	if backup.Command == "create" {
		dbID, err := ms.Master().QueryDBName2Id(ctx, dbName)
		if err != nil {
			return err
		}

		space, err := ms.Master().QuerySpaceByName(ctx, dbID, spaceName)
		if err != nil {
			return err
		}

		spaceJson, err := vjson.Marshal(space)
		if err != nil {
			log.Error("vjson.Marshal err: %v", err)
		}

		backupFileName := space.Name + ".schema"

		err = os.WriteFile(backupFileName, spaceJson, 0644)
		if err != nil {
			err := fmt.Errorf("error writing to file: %v", err)
			log.Error(err)
			return err
		}

		minioClient, err := minio.New(backup.S3Param.EndPoint, &minio.Options{
			Creds:  credentials.NewStaticV4(backup.S3Param.AccessKey, backup.S3Param.SecretKey, ""),
			Secure: backup.S3Param.UseSSL,
		})
		if err != nil {
			err = fmt.Errorf("failed to create minio client: %+v", err)
			log.Error(err)
			return err
		}
		bucketName := backup.S3Param.BucketName
		objectName := fmt.Sprintf("%s/%s/%s/%s.schema", clusterName, dbName, space.Name, space.Name)
		_, err = minioClient.FPutObject(context.Background(), bucketName, objectName, backupFileName, minio.PutObjectOptions{ContentType: "application/octet-stream"})
		if err != nil {
			err = fmt.Errorf("failed to backup space: %+v", err)
			log.Error(err)
			return err
		}
		log.Info("backup schema success, file is [%s]", backupFileName)
		os.Remove(backupFileName)
	} else if backup.Command == "restore" {
		dbID, err := ms.Master().QueryDBName2Id(ctx, dbName)
		if err != nil {
			return err
		}

		_, err = ms.Master().QuerySpaceByName(ctx, dbID, spaceName)
		if err == nil {
			err = fmt.Errorf("space duplicate")
			return err
		}
		minioClient, err := minio.New(backup.S3Param.EndPoint, &minio.Options{
			Creds:  credentials.NewStaticV4(backup.S3Param.AccessKey, backup.S3Param.SecretKey, ""),
			Secure: backup.S3Param.UseSSL,
		})
		if err != nil {
			err := fmt.Errorf("failed to create minio client: %+v", err)
			log.Error(err)
			return err
		}

		backupFileName := spaceName + ".schema"
		bucketName := backup.S3Param.BucketName
		objectName := fmt.Sprintf("%s/%s/%s/%s.schema", clusterName, dbName, spaceName, spaceName)
		err = minioClient.FGetObject(ctx, bucketName, objectName, backupFileName, minio.GetObjectOptions{})
		if err != nil {
			err := fmt.Errorf("failed to download file from S3: %+v", err)
			log.Error(err)
			return err
		}
		defer os.Remove(backupFileName)
		log.Info("downloaded backup file from S3: %s", backupFileName)

		spaceJson, err := os.ReadFile(backupFileName)
		if err != nil {
			err := fmt.Errorf("error read file:%v", err)
			log.Error(err)
			return err
		}

		log.Debug("%s", spaceJson)
		space := &entity.Space{}
		err = vjson.Unmarshal(spaceJson, space)
		if err != nil {
			err := fmt.Errorf("unmarshal file: %v", err)
			log.Error(err)
			return err
		}

		partitionNum := len(space.Partitions)
		space.Partitions = make([]*entity.Partition, 0)

		objectCh := minioClient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
			Recursive: true,
		})

		patitionMap := make(map[string]string, 0)
		for object := range objectCh {
			if object.Err != nil {
				fmt.Println(object.Err)
				continue
			}
			if strings.HasSuffix(object.Key, ".json.zst") {
				patitionMap[object.Key] = object.Key
			}
		}

		if len(patitionMap) != partitionNum {
			err = fmt.Errorf("oss partition num %d not equal schema %d", len(patitionMap), partitionNum)
			return err
		}
		if err := ms.createSpaceService(ctx, dbName, space); err != nil {
			log.Error("createSpaceService err: %v", err)
			return err
		}

		cfg, err := ms.GetEngineCfg(ctx, dbName, spaceName)
		if err != nil {
			log.Error("get engine config err: %s", err.Error())
			return err
		}

		err = ms.updateEngineConfig(ctx, space, cfg)
		if err != nil {
			log.Error("update engine config err: %s", err.Error())
			return err
		}
	}
	// get space info
	dbId, err := ms.Master().QueryDBName2Id(ctx, dbName)
	if err != nil {
		errutil.ThrowError(err)
	}

	space, err := ms.Master().QuerySpaceByName(ctx, dbId, spaceName)
	if err != nil {
		errutil.ThrowError(err)
	}
	if space == nil || space.Partitions == nil {
		return nil
	}
	// invoke all space nodeID
	part := 0
	for _, p := range space.Partitions {
		partition, err := ms.Master().QueryPartition(ctx, p.Id)
		if err != nil {
			log.Error(err)
			continue
		}
		if partition.Replicas != nil {
			for _, nodeID := range partition.Replicas {
				log.Debug("nodeID is [%+v],partition is [%+v], [%+v]", nodeID, partition.Id, partition.LeaderID)
				if nodeID != partition.LeaderID {
					continue
				}
				server, err := ms.Master().QueryServer(ctx, nodeID)
				errutil.ThrowError(err)
				log.Debug("invoke nodeID [%+v],address [%+v]", partition.Id, server.RpcAddr())
				backup.Part = part
				err = client.BackupSpace(server.RpcAddr(), backup, partition.Id)
				errutil.ThrowError(err)
				part++
			}
			if len(partition.Replicas) == 1 && partition.LeaderID == 0 {
				server, err := ms.Master().QueryServer(ctx, partition.Replicas[0])
				errutil.ThrowError(err)
				log.Debug("invoke nodeID [%+v],address [%+v]", partition.Id, server.RpcAddr())
				backup.Part = part
				err = client.BackupSpace(server.RpcAddr(), backup, partition.Id)
				errutil.ThrowError(err)
				part++
			}
		}
	}
	return nil
}

func (ms *masterService) ResourceLimitService(ctx context.Context, resourceLimit *entity.ResourceLimit) (err error) {
	spaces := make([]*entity.Space, 0)
	dbNames := make([]string, 0)
	if resourceLimit.DbName == nil && resourceLimit.SpaceName != nil {
		return vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("if space_name is set, db_name must be set"))
	}
	if resourceLimit.DbName != nil {
		dbNames = append(dbNames, *resourceLimit.DbName)
	}

	if len(dbNames) == 0 {
		dbs, err := ms.queryDBs(ctx)
		if err != nil {
			return err
		}
		dbNames = make([]string, len(dbs))
		for i, db := range dbs {
			dbNames[i] = db.Name
		}
	}

	for _, dbName := range dbNames {
		dbID, err := ms.Master().QueryDBName2Id(ctx, dbName)
		if err != nil {
			return err
		}
		if resourceLimit.SpaceName != nil {
			if space, err := ms.Master().QuerySpaceByName(ctx, dbID, *resourceLimit.SpaceName); err != nil {
				return err
			} else {
				spaces = append(spaces, space)
			}
		} else {
			if dbSpaces, err := ms.Master().QuerySpaces(ctx, dbID); err != nil {
				return err
			} else {
				spaces = append(spaces, dbSpaces...)
			}
		}
	}

	log.Debug("dbNames: %v, len(spaces): %d", dbNames, len(spaces))
	check := false
	for _, space := range spaces {
		for _, partition := range space.Partitions {
			for _, nodeID := range partition.Replicas {
				if server, err := ms.Master().QueryServer(ctx, nodeID); err != nil {
					return err
				} else {
					check = true
					err = client.ResourceLimit(server.RpcAddr(), resourceLimit, partition.Id)
					return err
				}
			}
		}
	}
	if !check {
		return vearchpb.NewError(vearchpb.ErrorEnum_INTERNAL_ERROR, fmt.Errorf("cluster is empty, no need to check resource limit"))
	}

	return nil
}

func (ms *masterService) updateSpaceResourceService(ctx context.Context, spaceResource *entity.SpacePartitionResource) (*entity.Space, error) {
	// it will lock cluster, to update space
	mutex := ms.Master().NewLock(ctx, entity.LockSpaceKey(spaceResource.DbName, spaceResource.SpaceName), time.Second*300)
	if err := mutex.Lock(); err != nil {
		return nil, err
	}
	defer func() {
		if err := mutex.Unlock(); err != nil {
			log.Error("failed to unlock space, the Error is:%v ", err)
		}
	}()

	dbId, err := ms.Master().QueryDBName2Id(ctx, spaceResource.DbName)
	if err != nil {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("failed to find database id according database name:%v,the Error is:%v ", spaceResource.DbName, err))
	}

	space, err := ms.Master().QuerySpaceByName(ctx, dbId, spaceResource.SpaceName)
	if err != nil {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("failed to find space according space name:%v,the Error is:%v ", spaceResource.SpaceName, err))
	}

	if space == nil {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("can not found space by name : %s", spaceResource.SpaceName))
	}

	if spaceResource.PartitionOperatorType != "" {
		if spaceResource.PartitionOperatorType != entity.Add && spaceResource.PartitionOperatorType != entity.Drop {
			return nil, vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("partition operator type should be %s or %s, but is %s", entity.Add, entity.Drop, spaceResource.PartitionOperatorType))
		}
		return ms.updateSpacePartitonRuleService(ctx, spaceResource, space)
	}

	// now only support update partition num
	if space.PartitionNum >= spaceResource.PartitionNum {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("paritition_num: %d now should greater than origin space partition_num: %d", spaceResource.PartitionNum, space.PartitionNum))
	}

	partitions := make([]*entity.Partition, 0)
	for i := space.PartitionNum; i < spaceResource.PartitionNum; i++ {
		partitionID, err := ms.Master().NewIDGenerate(ctx, entity.PartitionIdSequence, 1, 5*time.Second)

		if err != nil {
			return nil, err
		}

		partitions = append(partitions, &entity.Partition{
			Id:      entity.PartitionID(partitionID),
			SpaceId: space.Id,
			DBId:    space.DBId,
		})
		log.Debug("updateSpacePartitionNum Generate partition id %d", partitionID)
	}

	// find all servers for update space partition
	servers, err := ms.Master().QueryServers(ctx)
	if err != nil {
		return nil, err
	}

	// will get all exist partition
	serverPartitions, err := ms.filterAndSortServer(ctx, space, servers)
	if err != nil {
		return nil, err
	}

	if int(space.ReplicaNum) > len(serverPartitions) {
		return nil, fmt.Errorf("not enough PS , need replica %d but only has %d",
			int(space.ReplicaNum), len(serverPartitions))
	}

	// pick servers for space
	var paddrs [][]string
	for i := 0; i < len(partitions); i++ {
		if addrs, err := ms.selectServersForPartition(servers, serverPartitions, space.ReplicaNum, partitions[i]); err != nil {
			return nil, err
		} else {
			paddrs = append(paddrs, addrs)
		}
	}

	log.Debug("updateSpacePartitionNum origin paritionNum %d, serverPartitions %v, paddrs %v", space.PartitionNum, serverPartitions, paddrs)

	// when create partition, new partition id will be stored in server partition cache
	space.PartitionNum = spaceResource.PartitionNum
	space.Partitions = append(space.Partitions, partitions...)

	var errChain = make(chan error, 1)
	// send create partition for new
	for i := 0; i < len(partitions); i++ {
		go func(addrs []string, partition *entity.Partition) {
			//send request for all server
			defer func() {
				if r := recover(); r != nil {
					err := vearchpb.NewError(vearchpb.ErrorEnum_INTERNAL_ERROR, fmt.Errorf("create partition err: %v ", r))
					errChain <- err
					log.Error(err.Error())
				}
			}()
			for _, addr := range addrs {
				if err := client.CreatePartition(addr, space, partition.Id); err != nil {
					err := vearchpb.NewError(vearchpb.ErrorEnum_INTERNAL_ERROR, fmt.Errorf("create partition err: %s ", err.Error()))
					errChain <- err
					log.Error(err.Error())
				}
			}
		}(paddrs[i], partitions[i])
	}

	// check all partition is ok
	for i := 0; i < len(partitions); i++ {
		times := 0
		for {
			times++
			select {
			case err := <-errChain:
				return nil, err
			case <-ctx.Done():
				return nil, vearchpb.NewError(vearchpb.ErrorEnum_INTERNAL_ERROR, fmt.Errorf("update space has error"))
			default:
			}

			partition, err := ms.Master().QueryPartition(ctx, partitions[i].Id)
			if times%5 == 0 {
				log.Debug("updateSpacePartitionNum check the partition:%d status", partitions[i].Id)
			}
			if err != nil && vearchpb.NewError(vearchpb.ErrorEnum_INTERNAL_ERROR, err).GetError().Code != vearchpb.ErrorEnum_PARTITION_NOT_EXIST {
				return nil, err
			}
			if partition != nil {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
	}

	//update space
	width := math.MaxUint32 / space.PartitionNum
	for i := 0; i < space.PartitionNum; i++ {
		space.Partitions[i].Slot = entity.SlotID(i * width)
	}
	log.Debug("updateSpacePartitionNum space version %d, partition_num %d", space.Version, space.PartitionNum)

	if err := ms.updateSpace(ctx, space); err != nil {
		return nil, err
	} else {
		return space, nil
	}
}

func (ms *masterService) updateSpacePartitonRuleService(ctx context.Context, spaceResource *entity.SpacePartitionResource, space *entity.Space) (*entity.Space, error) {
	if spaceResource.PartitionOperatorType == entity.Drop {
		if spaceResource.PartitionName == "" {
			return nil, vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("partition name is empty"))
		}
		found := false
		for _, range_rule := range space.PartitionRule.Ranges {
			if range_rule.Name == spaceResource.PartitionName {
				found = true
				break
			}
		}
		if !found {
			return nil, vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("partition name %s not exist", spaceResource.PartitionName))
		}
		new_partitions := make([]*entity.Partition, 0)
		for _, partition := range space.Partitions {
			if partition.Name != spaceResource.PartitionName {
				new_partitions = append(new_partitions, partition)
			} else {
				// delete parition and partitionKey
				for _, replica := range partition.Replicas {
					if server, err := ms.Master().QueryServer(ctx, replica); err != nil {
						log.Error("query partition:[%d] for replica:[%s] has err:[%s]", partition.Id, replica, err.Error())
					} else {
						if err := client.DeletePartition(server.RpcAddr(), partition.Id); err != nil {
							log.Error("delete partition:[%d] for server:[%s] has err:[%s]", partition.Id, server.RpcAddr(), err.Error())
						}
					}
				}
				err := ms.Master().Delete(ctx, entity.PartitionKey(partition.Id))
				if err != nil {
					return nil, err
				}
			}
		}
		space.Partitions = new_partitions
		new_range_rules := make([]entity.Range, 0)
		for _, range_rule := range space.PartitionRule.Ranges {
			if range_rule.Name != spaceResource.PartitionName {
				new_range_rules = append(new_range_rules, range_rule)
			}
		}
		space.PartitionRule.Ranges = new_range_rules
	}

	if spaceResource.PartitionOperatorType == entity.Add {
		if spaceResource.PartitionRule == nil {
			return nil, vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("partition rule is empty"))
		}
		_, err := space.PartitionRule.RangeIsSame(spaceResource.PartitionRule.Ranges)
		if err != nil {
			return nil, err
		}

		// find all servers for update space partition
		servers, err := ms.Master().QueryServers(ctx)
		if err != nil {
			return nil, err
		}

		// will get all exist partition
		serverPartitions, err := ms.filterAndSortServer(ctx, space, servers)
		if err != nil {
			return nil, err
		}

		if int(space.ReplicaNum) > len(serverPartitions) {
			return nil, fmt.Errorf("not enough PS , need replica %d but only has %d",
				int(space.ReplicaNum), len(serverPartitions))
		}

		partitions := make([]*entity.Partition, 0)
		for _, r := range spaceResource.PartitionRule.Ranges {
			for j := 0; j < space.PartitionNum; j++ {
				partitionID, err := ms.Master().NewIDGenerate(ctx, entity.PartitionIdSequence, 1, 5*time.Second)

				if err != nil {
					return nil, err
				}

				partitions = append(partitions, &entity.Partition{
					Id:      entity.PartitionID(partitionID),
					Name:    r.Name,
					SpaceId: space.Id,
					DBId:    space.DBId,
				})
				log.Debug("updateSpacePartitionrule Generate partition id %d", partitionID)
			}
		}
		space.PartitionRule.Ranges, err = space.PartitionRule.AddRanges(spaceResource.PartitionRule.Ranges)
		if err != nil {
			return nil, err
		}
		log.Debug("updateSpacePartitionrule partition rule %v, add rule %v", space.PartitionRule, spaceResource.PartitionRule)

		// pick servers for space
		var paddrs [][]string
		for i := 0; i < len(partitions); i++ {
			if addrs, err := ms.selectServersForPartition(servers, serverPartitions, space.ReplicaNum, partitions[i]); err != nil {
				return nil, err
			} else {
				paddrs = append(paddrs, addrs)
			}
		}

		log.Debug("updateSpacePartitionrule paritionNum %d, serverPartitions %v, paddrs %v", space.PartitionNum, serverPartitions, paddrs)

		// when create partition, new partition id will be stored in server partition cache
		space.Partitions = append(space.Partitions, partitions...)

		var errChain = make(chan error, 1)
		// send create partition for new
		for i := 0; i < len(partitions); i++ {
			go func(addrs []string, partition *entity.Partition) {
				//send request for all server
				defer func() {
					if r := recover(); r != nil {
						err := vearchpb.NewError(vearchpb.ErrorEnum_INTERNAL_ERROR, fmt.Errorf("create partition err: %v ", r))
						errChain <- err
						log.Error(err.Error())
					}
				}()
				for _, addr := range addrs {
					if err := client.CreatePartition(addr, space, partition.Id); err != nil {
						err := vearchpb.NewError(vearchpb.ErrorEnum_INTERNAL_ERROR, fmt.Errorf("create partition err: %s ", err.Error()))
						errChain <- err
						log.Error(err.Error())
					}
				}
			}(paddrs[i], partitions[i])
		}
		// check all partition is ok
		for i := 0; i < len(partitions); i++ {
			times := 0
			for {
				times++
				select {
				case err := <-errChain:
					return nil, err
				case <-ctx.Done():
					return nil, vearchpb.NewError(vearchpb.ErrorEnum_INTERNAL_ERROR, fmt.Errorf("update space has error"))
				default:
				}

				partition, err := ms.Master().QueryPartition(ctx, partitions[i].Id)
				if times%5 == 0 {
					log.Debug("updateSpacePartitionNum check the partition:%d status", partitions[i].Id)
				}
				if err != nil && vearchpb.NewError(vearchpb.ErrorEnum_INTERNAL_ERROR, err).GetError().Code != vearchpb.ErrorEnum_PARTITION_NOT_EXIST {
					return nil, err
				}
				if partition != nil {
					break
				}
				time.Sleep(50 * time.Millisecond)
			}
		}
	}

	space.PartitionRule.Partitions = len(space.PartitionRule.Ranges)
	//update space
	width := math.MaxUint32 / (space.PartitionNum * space.PartitionRule.Partitions)
	for i := 0; i < space.PartitionNum*space.PartitionRule.Partitions; i++ {
		space.Partitions[i].Slot = entity.SlotID(i * width)
	}

	if err := ms.updateSpace(ctx, space); err != nil {
		return nil, err
	} else {
		return space, nil
	}
}

func (ms *masterService) ChangeMember(ctx context.Context, cm *entity.ChangeMember) error {
	partition, err := ms.Master().QueryPartition(ctx, cm.PartitionID)
	if err != nil {
		log.Error(err)
		return err
	}

	space, err := ms.Master().QuerySpaceByID(ctx, partition.DBId, partition.SpaceId)
	if err != nil {
		return err
	}

	dbName, err := ms.Master().QueryDBId2Name(ctx, space.DBId)
	if err != nil {
		return err
	}

	spacePartition := space.GetPartition(cm.PartitionID)

	if cm.Method != proto.ConfRemoveNode {
		for _, nodeID := range spacePartition.Replicas {
			if nodeID == cm.NodeID {
				log.Error("partition:[%d] already on server:[%d] in replicas:[%v], space:%v", cm.PartitionID, cm.NodeID, spacePartition.Replicas, space)
				return vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("partition:[%d] already on server:[%d] in replicas:[%v]", cm.PartitionID, cm.NodeID, spacePartition.Replicas))
			}
		}
		spacePartition.Replicas = append(spacePartition.Replicas, cm.NodeID)
	} else {
		tempIDs := make([]entity.NodeID, 0, len(spacePartition.Replicas)-1)

		for _, id := range spacePartition.Replicas {
			if id != cm.NodeID {
				tempIDs = append(tempIDs, id)
			}
		}
		spacePartition.Replicas = tempIDs
	}

	retryTimes := 10
	masterNode, err := ms.Master().QueryServer(ctx, partition.LeaderID)
	for err != nil {
		partition, err = ms.Master().QueryPartition(ctx, cm.PartitionID)
		if err != nil {
			log.Error(err)
			return err
		}
		masterNode, err = ms.Master().QueryServer(ctx, partition.LeaderID)
		if err != nil {
			retryTimes--
			if retryTimes == 0 {
				return err
			}
			time.Sleep(10 * time.Second)
		}
	}
	log.Info("masterNode is [%+v], cm is [%+v] ", masterNode, cm)

	var targetNode *entity.Server
	targetNode, err = ms.Master().QueryServer(ctx, cm.NodeID)
	if err != nil {
		if cm.Method == proto.ConfRemoveNode {
			// maybe node is crashed
			targetNode = nil
		} else {
			return err
		}
	}
	log.Info("targetNode is [%+v], cm is [%+v] ", targetNode, cm)

	if !client.IsLive(masterNode.RpcAddr()) {
		return vearchpb.NewError(vearchpb.ErrorEnum_PARTITION_SERVER_ERROR, fmt.Errorf("server:[%d] addr:[%s] can not connect ", cm.NodeID, masterNode.RpcAddr()))
	}

	if cm.Method == proto.ConfAddNode && targetNode != nil {
		if err := client.CreatePartition(targetNode.RpcAddr(), space, cm.PartitionID); err != nil {
			return vearchpb.NewError(vearchpb.ErrorEnum_INTERNAL_ERROR, fmt.Errorf("create partiiton has err:[%s] addr:[%s]", err.Error(), targetNode.RpcAddr()))
		}
	} else if cm.Method == proto.ConfRemoveNode {

	} else {
		return vearchpb.NewError(vearchpb.ErrorEnum_INTERNAL_ERROR, fmt.Errorf("change member only support add:[%d] remove:[%d] not support:[%d]", proto.ConfAddNode, proto.ConfRemoveNode, cm.Method))
	}

	if err := client.ChangeMember(masterNode.RpcAddr(), cm); err != nil {
		return err
	}
	if cm.Method == proto.ConfRemoveNode && targetNode != nil && client.IsLive(targetNode.RpcAddr()) {
		if err := client.DeleteReplica(targetNode.RpcAddr(), cm.PartitionID); err != nil {
			return vearchpb.NewError(vearchpb.ErrorEnum_INTERNAL_ERROR, fmt.Errorf("delete partiiton has err:[%s] addr:[%s]", err.Error(), targetNode.RpcAddr()))
		}
	}

	if _, err := ms.updateSpaceService(ctx, dbName, space.Name, space); err != nil {
		return err
	}
	log.Info("update space: %v", space)
	return nil
}

func (ms *masterService) ChangeMembers(ctx context.Context, cms *entity.ChangeMembers) error {
	for _, partitionId := range cms.PartitionIDs {
		cm := &entity.ChangeMember{
			PartitionID: uint32(partitionId),
			NodeID:      cms.NodeID,
			Method:      cms.Method,
		}
		if err := ms.ChangeMember(ctx, cm); err != nil {
			return err
		}
	}
	return nil
}

// recover fail node
func (ms *masterService) RecoverFailServer(ctx context.Context, rs *entity.RecoverFailServer) (e error) {
	// panic process
	defer errutil.CatchError(&e)
	// get fail server info
	targetFailServer := ms.Master().QueryServerByIPAddr(ctx, rs.FailNodeAddr)
	log.Debug("targetFailServer is %s", targetFailServer)
	// get new server info
	newServer := ms.Master().QueryServerByIPAddr(ctx, rs.NewNodeAddr)
	log.Debug("newServer is %s", newServer)
	if newServer.ID <= 0 || targetFailServer.ID <= 0 {
		return vearchpb.NewError(vearchpb.ErrorEnum_PARTITION_SERVER_ERROR, fmt.Errorf("newServer or targetFailServer is nil"))
	}

	for _, pid := range targetFailServer.Node.PartitionIds {
		cm := &entity.ChangeMember{}
		cm.Method = proto.ConfAddNode
		cm.NodeID = newServer.ID
		cm.PartitionID = pid
		if e = ms.ChangeMember(ctx, cm); e != nil {
			e = fmt.Errorf("ChangePartitionMember failed [%+v], err is %s ", cm, e.Error())
			log.Error(e)
			return e
		}
		log.Info("ChangePartitionMember add [%+v] success,", cm)
		cm.Method = proto.ConfRemoveNode
		cm.NodeID = targetFailServer.ID
		cm.PartitionID = pid
		if e = ms.ChangeMember(ctx, cm); e != nil {
			e = fmt.Errorf("ChangePartitionMember failed [%+v], err is %s ", cm, e.Error())
			log.Error(e)
			return e
		}
		log.Info("ChangePartitionMember remove [%+v]", cm)
	}
	// if success, remove from failServer
	ms.Master().TryRemoveFailServer(ctx, targetFailServer.Node)

	return nil
}

// get servers belong this db
func (ms *masterService) DBServers(ctx context.Context, dbName string) (servers []*entity.Server, err error) {
	defer errutil.CatchError(&err)
	// get all servers
	servers, err = ms.Master().QueryServers(ctx)
	errutil.ThrowError(err)

	db, err := ms.queryDBService(ctx, dbName)
	if err != nil {
		return nil, err
	}
	// get private server
	if len(db.Ps) > 0 {
		privateServer := make([]*entity.Server, 0)
		for _, ps := range db.Ps {
			for _, s := range servers {
				if ps == s.Ip {
					privateServer = append(privateServer, s)
				}
			}
		}
		return privateServer, nil
	}
	return servers, nil
}

// Check if partition.Id is in s.PartitionIds
func isPartitionIdInSlice[T comparable](partitionId T, partitionIds []T) (bool, int) {
	for i, v := range partitionIds {
		if v == partitionId {
			return true, i
		}
	}
	return false, -1
}

// change replicas, add or delete
func (ms *masterService) ChangeReplica(ctx context.Context, dbModify *entity.DBModify) (e error) {
	// panic process
	defer errutil.CatchError(&e)
	// query server
	servers, err := ms.DBServers(ctx, dbModify.DbName)
	errutil.ThrowError(err)
	// generate change servers
	dbID, err := ms.Master().QueryDBName2Id(ctx, dbModify.DbName)
	errutil.ThrowError(err)
	space, err := ms.Master().QuerySpaceByName(ctx, dbID, dbModify.SpaceName)
	errutil.ThrowError(err)
	if dbModify.Method == proto.ConfAddNode && (int(space.ReplicaNum)+1) > len(servers) {
		err := vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf("ReplicaNum [%d] exceeds server size [%d]",
			int(space.ReplicaNum)+1, len(servers)))
		return err
	}
	// change space replicas of partition, add or delete one
	changeServer := make([]*entity.ChangeMember, 0)
	for _, partition := range space.Partitions {
		// sort servers, low to high
		sort.Slice(servers, func(i, j int) bool {
			return len(servers[i].PartitionIds) < len(servers[j].PartitionIds)
		})
		// change server
		for _, s := range servers {
			if dbModify.Method == proto.ConfAddNode {
				exist, _ := isPartitionIdInSlice(s.ID, partition.Replicas)
				if !exist {
					// server doesn't contain this partition, then create it
					cm := &entity.ChangeMember{PartitionID: partition.Id, NodeID: s.ID, Method: dbModify.Method}
					changeServer = append(changeServer, cm)
					s.PartitionIds = append(s.PartitionIds, partition.Id)
					break
				}
			} else if dbModify.Method == proto.ConfRemoveNode {
				exist, index := isPartitionIdInSlice(partition.Id, s.PartitionIds)
				if exist {
					// server contains this partition, then remove it
					cm := &entity.ChangeMember{PartitionID: partition.Id, NodeID: s.ID, Method: dbModify.Method}
					changeServer = append(changeServer, cm)
					s.PartitionIds = append(s.PartitionIds[:index], s.PartitionIds[index+1:]...)
					break
				}
			}
		}
	}
	log.Info("need to change partition is [%+v] ", vjson.ToJsonString(changeServer))
	// sleep time
	sleepTime := config.Conf().PS.RaftHeartbeatInterval
	// change partition
	for _, cm := range changeServer {
		if e = ms.ChangeMember(ctx, cm); e != nil {
			info := fmt.Sprintf("changing partition member [%+v] failed, error is %s ", cm, e)
			log.Error(info)
			panic(fmt.Errorf(info))
		}
		log.Info("changing partition member [%+v] succeeded ", cm)
		if dbModify.Method == proto.ConfRemoveNode {
			time.Sleep(time.Duration(sleepTime*10) * time.Millisecond)
			log.Info("remove partition sleep [%+d] milliseconds", sleepTime)
		}
	}
	log.Info("all partition member changes succeeded")

	// update space ReplicaNum, it will lock cluster, to create space
	mutex := ms.Master().NewLock(ctx, entity.LockSpaceKey(dbModify.DbName, dbModify.SpaceName), time.Second*300)
	if _, err := mutex.TryLock(); err != nil {
		errutil.ThrowError(err)
	}
	defer func() {
		if err := mutex.Unlock(); err != nil {
			log.Error("failed to unlock space, the error is: %v ", err)
		}
	}()
	space, err = ms.Master().QuerySpaceByName(ctx, dbID, dbModify.SpaceName)
	errutil.ThrowError(err)
	if dbModify.Method == proto.ConfAddNode {
		space.ReplicaNum = space.ReplicaNum + 1
	} else if dbModify.Method == proto.ConfRemoveNode {
		space.ReplicaNum = space.ReplicaNum - 1
	}
	err = ms.updateSpace(ctx, space)
	log.Info("updateSpace space [%+v] succeeded", space)
	errutil.ThrowError(err)
	return e
}

func (ms *masterService) IsExistNode(ctx context.Context, id entity.NodeID, ip string) error {
	values, err := ms.Master().Get(ctx, entity.ServerKey(id))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("look up key[%s] in etcd failed", entity.ServerKey(id)))
	}
	if values == nil {
		return nil
	}
	server := &entity.Server{}
	err = vjson.Unmarshal(values, server)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("parse key[%s] in etcd failed", entity.ServerKey(id)))
	}
	if server.Ip != ip {
		return errors.Errorf("node id[%d] has register on ip[%s]", server.ID, server.Ip)
	}
	return nil
}

func (ms *masterService) partitionInfo(ctx context.Context, dbName string, spaceName string, detail string) ([]map[string]interface{}, error) {
	dbNames := make([]string, 0)
	if dbName != "" {
		dbNames = strings.Split(dbName, ",")
	}

	if len(dbNames) == 0 {
		dbs, err := ms.queryDBs(ctx)
		if err != nil {
			return nil, err
		}
		dbNames = make([]string, len(dbs))
		for i, db := range dbs {
			dbNames[i] = db.Name
		}
	}

	color := []string{"green", "yellow", "red"}

	spaceNames := make([]string, 0)
	if spaceName != "" {
		spaceNames = strings.Split(spaceName, ",")
	}

	detail_info := false
	if detail == "true" {
		detail_info = true
	}

	resultInsideDbs := make([]map[string]interface{}, 0)
	for i := range dbNames {
		dbName := dbNames[i]
		var errors []string

		resultDb := make(map[string]interface{})
		resultDb["db_name"] = dbName

		dbId, err := ms.Master().QueryDBName2Id(ctx, dbName)
		if err != nil {
			errors = append(errors, "db: "+dbName+" find dbID err: "+err.Error())
			resultDb["errors"] = errors
			resultInsideDbs = append(resultInsideDbs, resultDb)
			continue
		}

		spaces, err := ms.Master().QuerySpaces(ctx, dbId)
		if err != nil {
			errors = append(errors, "db: "+dbName+" find spaces err: "+err.Error())
			resultDb["errors"] = errors
			resultInsideDbs = append(resultInsideDbs, resultDb)
			continue
		}

		dbStatus := 0

		resultInsideSpaces := make([]*entity.SpaceInfo, 0, len(spaces))
		if len(spaceNames) == 0 {
			for _, space := range spaces {
				spaceName := space.Name

				spaceInfo := &entity.SpaceInfo{Name: spaceName, DbName: dbName, ReplicaNum: space.ReplicaNum, PartitionNum: space.PartitionNum}
				spaceStatus, err := ms.describeSpaceService(ctx, space, spaceInfo, detail_info)
				if err != nil {
					log.Error(err.Error())
					errors = append(errors, "db: "+dbName+" space: "+spaceName+" describe err: "+err.Error())
					continue
				}
				resultInsideSpaces = append(resultInsideSpaces, spaceInfo)

				if spaceStatus > dbStatus {
					dbStatus = spaceStatus
				}
			}
		} else {
			for _, spaceName := range spaceNames {
				index := -1
				for i, space := range spaces {
					if space.Name == spaceName {
						index = i
						break
					}
				}

				if index < 0 {
					msg := fmt.Sprintf("db: %s space: %s not found", dbName, spaceName)
					errors = append(errors, msg)
					continue
				}

				spaceInfo := &entity.SpaceInfo{Name: spaceName, DbName: dbName, ReplicaNum: spaces[index].ReplicaNum, PartitionNum: spaces[index].PartitionNum}
				spaceStatus, err := ms.describeSpaceService(ctx, spaces[index], spaceInfo, detail_info)
				if err != nil {
					log.Error(err.Error())
					errors = append(errors, "db: "+dbName+" space: "+spaceName+" describe err: "+err.Error())
					continue
				}
				resultInsideSpaces = append(resultInsideSpaces, spaceInfo)

				if spaceStatus > dbStatus {
					dbStatus = spaceStatus
				}
			}
		}

		docNum := uint64(0)
		// TODO: get size
		size := int64(0)
		for _, s := range resultInsideSpaces {
			docNum += cast.ToUint64(s.DocNum)
		}
		resultDb["space_num"] = len(spaces)
		resultDb["doc_num"] = docNum
		resultDb["size"] = size
		resultDb["spaces"] = resultInsideSpaces
		resultDb["status"] = color[dbStatus]
		resultDb["errors"] = errors
		resultInsideDbs = append(resultInsideDbs, resultDb)
	}

	return resultInsideDbs, nil
}

func (ms *masterService) statsService(ctx context.Context) ([]*mserver.ServerStats, error) {
	servers, err := ms.Master().QueryServers(ctx)
	if err != nil {
		return nil, err
	}

	statsChan := make(chan *mserver.ServerStats, len(servers))

	for _, s := range servers {
		go func(s *entity.Server) {
			defer func() {
				if r := recover(); r != nil {
					statsChan <- mserver.NewErrServerStatus(s.RpcAddr(), errors.New(cast.ToString(r)))
				}
			}()
			statsChan <- client.ServerStats(s.RpcAddr())
		}(s)
	}

	result := make([]*mserver.ServerStats, 0, len(servers))

	for {
		select {
		case s := <-statsChan:
			result = append(result, s)
		case <-ctx.Done():
			return nil, vearchpb.NewError(vearchpb.ErrorEnum_TIMEOUT, nil)
		default:
			time.Sleep(time.Millisecond * 10)
			if len(result) >= len(servers) {
				close(statsChan)
				goto out
			}
		}
	}

out:

	return result, nil
}

func (ms *masterService) addMemberService(ctx context.Context, peerAddrs []string) (resp *clientv3.MemberAddResponse, err error) {
	resp, err = ms.Master().MemberAdd(ctx, peerAddrs)
	if err != nil {
		log.Error("add masters member err:%s", err.Error())
		return nil, err
	}
	return resp, nil
}

func (ms *masterService) removeMemberService(ctx context.Context, master *entity.MemberInfoRequest) (resp *clientv3.MemberRemoveResponse, err error) {
	membersResp, err := ms.Master().MemberList(ctx)
	if err != nil {
		log.Error("master member list err:%s", err.Error())
		return nil, err
	}
	if len(membersResp.Members) <= 1 {
		msg := fmt.Sprintf("master member only have %d, cann't remove member now", len(membersResp.Members))
		log.Error(msg)
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf(msg))
	}

	found := false
	for _, member := range membersResp.Members {
		if member.ID == master.ID && member.Name == master.Name {
			found = true
			break
		}
	}
	if !found {
		msg := fmt.Sprintf("master member name:%s id:%d not found", master.Name, master.ID)
		log.Error(msg)
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf(msg))
	}
	resp, err = ms.Master().MemberRemove(ctx, master.ID)
	if err != nil {
		msg := fmt.Sprintf("remove master member err:%s", err.Error())
		log.Error(msg)
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_PARAM_ERROR, fmt.Errorf(msg))
	}

	err = ms.Master().STM(context.Background(), func(stm concurrency.STM) error {
		stm.Del(entity.MasterMemberKey(master.ID))
		return nil
	})
	if err != nil {
		log.Error("del masters err:%s", err.Error())
		return nil, err
	}
	// TODO other master client should also do member sync
	err = ms.Master().MemberSync(ctx)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
