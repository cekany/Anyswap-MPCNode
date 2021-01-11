/*
 *  Copyright (C) 2018-2019  Fusion Foundation Ltd. All rights reserved.
 *  Copyright (C) 2018-2019  caihaijun@fusion.org
 *
 *  This library is free software; you can redistribute it and/or
 *  modify it under the Apache License, Version 2.0.
 *
 *  This library is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 *
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

package dcrm 

import (
    "github.com/anyswap/Anyswap-MPCNode/internal/common"
    "github.com/anyswap/Anyswap-MPCNode/internal/common/fdlimit"
    "github.com/anyswap/Anyswap-MPCNode/ethdb"
    "time"
    "sync"
    "github.com/anyswap/Anyswap-MPCNode/p2p/discover"
)

var (
	LdbPubKeyData  = common.NewSafeMap(10) //make(map[string][]byte)
	PubKeyDataChan = make(chan KeyData, 2000)
	SkU1Chan = make(chan KeyData, 2000)
	Bip32CChan = make(chan KeyData, 2000)
	cache = (75*1024)/1000 
	handles = makeDatabaseHandles()
	
	lock                     sync.Mutex
	db *ethdb.LDBDatabase
	dbsk *ethdb.LDBDatabase
	dbbip32 *ethdb.LDBDatabase
)

func makeDatabaseHandles() int {
     limit, err := fdlimit.Current()
     if err != nil {
	     //Fatalf("Failed to retrieve file descriptor allowance: %v", err)
	     common.Info("Failed to retrieve file descriptor allowance: " + err.Error())
	     return 0
     }
     if limit < 2048 {
	     if err := fdlimit.Raise(2048); err != nil {
		     //Fatalf("Failed to raise file descriptor allowance: %v", err)
		     common.Info("Failed to raise file descriptor allowance: " + err.Error())
	     }
     }
     if limit > 2048 { // cap database file descriptors even if more is available
	     limit = 2048
     }
     return limit / 2 // Leave half for networking and other stuff
}

func GetBip32CFromLocalDb(key string) []byte {
	lock.Lock()
	/*if dbbip32 == nil {
	    common.Debug("=====================GetSkU1FromLocalDb, dbsk is nil =====================")
	    dir := GetSkU1Dir()
	    ////////
	    dbsktmp, err := ethdb.NewLDBDatabase(dir, cache, handles)
	    //bug
	    if err != nil {
		    for i := 0; i < 100; i++ {
			    dbsktmp, err = ethdb.NewLDBDatabase(dir, cache, handles)
			    if err == nil {
				    break
			    }

			    time.Sleep(time.Duration(1000000))
		    }
	    }
	    if err != nil {
		dbsk = nil
	    } else {
		dbsk = dbsktmp
	    }

		lock.Unlock()
		return nil
	}*/

	da, err := dbbip32.Get([]byte(key))
	if err != nil {
	    dir := GetBip32CDir()
	    ////////
	    dbbiptmp, err := ethdb.NewLDBDatabase(dir, cache, handles)
	    //bug
	    if err != nil {
		    for i := 0; i < 100; i++ {
			    dbbiptmp, err = ethdb.NewLDBDatabase(dir, cache, handles)
			    if err == nil {
				    break
			    }

			    time.Sleep(time.Duration(1000000))
		    }
	    }
	    if err != nil {
	    } else {
		dbbip32 = dbbiptmp
	    }

	    da, err = dbbip32.Get([]byte(key))
	    if err != nil {
		lock.Unlock()
		return nil
	    }
	    
	    bip32c,err := DecryptMsg(string(da))
	    if err != nil {
		lock.Unlock()
		return da //TODO ,tmp code 
		//return nil
	    }

	    lock.Unlock()
	    return []byte(bip32c)
	}

	bip32c,err := DecryptMsg(string(da))
	if err != nil {
	    lock.Unlock()
	    return da //TODO ,tmp code 
	    //return nil
	}

	lock.Unlock()
	return []byte(bip32c)
}

func SaveBip32CToDb() {
	for {
		kd := <-Bip32CChan
		if dbbip32 != nil {
		    cm,err := EncryptMsg(kd.Data,cur_enode)
		    if err != nil {
			Bip32CChan <- kd
			continue	
		    }

		    err = dbbip32.Put(kd.Key, []byte(cm))
		    if err != nil {
			dir := GetBip32CDir()
			dbbiptmp, err := ethdb.NewLDBDatabase(dir, cache, handles)
			//bug
			if err != nil {
				for i := 0; i < 100; i++ {
					dbbiptmp, err = ethdb.NewLDBDatabase(dir, cache, handles)
					if err == nil {
						break
					}

					time.Sleep(time.Duration(1000000))
				}
			}
			if err != nil {
			} else {
			    dbbip32 = dbbiptmp
			    err = dbbip32.Put(kd.Key, []byte(cm))
			    if err != nil {
				Bip32CChan <- kd
				continue
			    }
			}

		    }
		//	db.Close()
		} else {
			Bip32CChan <- kd
		}

		time.Sleep(time.Duration(1000000)) //na, 1 s = 10e9 na
	    }
}

func GetBip32CDir() string {
	dir := common.DefaultDataDir()
	dir += "/dcrmdata/bip32" + cur_enode
	return dir
}

func GetSkU1FromLocalDb(key string) []byte {
	lock.Lock()
	if dbsk == nil {
	    common.Debug("=====================GetSkU1FromLocalDb, dbsk is nil =====================")
	    dir := GetSkU1Dir()
	    ////////
	    dbsktmp, err := ethdb.NewLDBDatabase(dir, cache, handles)
	    //bug
	    if err != nil {
		    for i := 0; i < 100; i++ {
			    dbsktmp, err = ethdb.NewLDBDatabase(dir, cache, handles)
			    if err == nil {
				    break
			    }

			    time.Sleep(time.Duration(1000000))
		    }
	    }
	    if err != nil {
		dbsk = nil
	    } else {
		dbsk = dbsktmp
	    }

		lock.Unlock()
		return nil
	}

	da, err := dbsk.Get([]byte(key))
	if err != nil {
	    dir := GetSkU1Dir()
	    ////////
	    dbsktmp, err := ethdb.NewLDBDatabase(dir, cache, handles)
	    //bug
	    if err != nil {
		    for i := 0; i < 100; i++ {
			    dbsktmp, err = ethdb.NewLDBDatabase(dir, cache, handles)
			    if err == nil {
				    break
			    }

			    time.Sleep(time.Duration(1000000))
		    }
	    }
	    if err != nil {
		//dbsk = nil
	    } else {
		dbsk = dbsktmp
	    }

	    da, err = dbsk.Get([]byte(key))
	    if err != nil {
		lock.Unlock()
		return nil
	    }
	    
	    sk,err := DecryptMsg(string(da))
	    if err != nil {
		lock.Unlock()
		return da //TODO ,tmp code 
		//return nil
	    }

	    lock.Unlock()
	    return []byte(sk)
	}

	sk,err := DecryptMsg(string(da))
	if err != nil {
	    lock.Unlock()
	    return da //TODO ,tmp code 
	    //return nil
	}

	lock.Unlock()
	return []byte(sk)
}

func GetPubKeyDataValueFromDb(key string) []byte {
	lock.Lock()
	/*if db == nil {
	    dir := GetDbDir()
	    ////////
	    dbtmp, err := ethdb.NewLDBDatabase(dir, cache, handles)
	    //bug
	    if err != nil {
		    for i := 0; i < 100; i++ {
			    dbtmp, err = ethdb.NewLDBDatabase(dir, cache, handles)
			    if err == nil {
				    break
			    }

			    time.Sleep(time.Duration(1000000))
		    }
	    }
	    if err != nil {
	    common.Debug("===================GetPubKeyDataValueFromDb, db is nil and re-get, ===================","err",err,"dir",dir,"key",key)
		lock.Unlock()
		return nil
	    } else {
		db = dbtmp
		da, err := db.Get([]byte(key))
		if err != nil {
	    common.Debug("===================GetPubKeyDataValueFromDb, db is nil and re-get success,but get data fail ===================","err",err,"dir",dir,"key",key)
		    lock.Unlock()
		    return nil
		}

		lock.Unlock()
		return da
	    }
	}
	*/

	if db == nil {
	    lock.Unlock()
	    return nil
 	}

	da, err := db.Get([]byte(key))
	if err != nil {
	    common.Info("===================GetPubKeyDataValueFromDb,get data fail===================","err",err,"key",key)

	    /*dir := GetDbDir()
	    ////////
	    dbtmp, err := ethdb.NewLDBDatabase(dir, cache, handles)
	    if err != nil {
		    for i := 0; i < 100; i++ {
			    dbtmp, err = ethdb.NewLDBDatabase(dir, cache, handles)
			    if err == nil {
				    break
			    }

			    time.Sleep(time.Duration(1000000))
		    }
	    }
	    if err != nil {
		lock.Unlock()
		return nil
	    } else {
		db = dbtmp
		da, err := db.Get([]byte(key))
		if err != nil {
		    lock.Unlock()
		    return nil
		}

		lock.Unlock()
		return da
	    }*/

	    lock.Unlock()
	    return nil
	}

	lock.Unlock()
	return da
}

type KeyData struct {
	Key  []byte
	Data string
}

func SavePubKeyDataToDb() {
	for {
		kd := <-PubKeyDataChan
		if db != nil {
		    if kd.Data == "CLEAN" {
			err := db.Delete(kd.Key)
			if err != nil {
				common.Info("=================SavePubKeyDataToDb, db is not nil and delete fail ===============","key",kd.Key)
			}
		    } else {
			err := db.Put(kd.Key, []byte(kd.Data))
			if err != nil {
				common.Info("=================SavePubKeyDataToDb, db is not nil and save fail ===============","key",kd.Key)
			    dir := GetDbDir()
			    dbtmp, err := ethdb.NewLDBDatabase(dir, cache, handles)
			    //bug
			    if err != nil {
				    for i := 0; i < 100; i++ {
					    dbtmp, err = ethdb.NewLDBDatabase(dir, cache, handles)
					    if err == nil {
						    break
					    }

					    time.Sleep(time.Duration(1000000))
				    }
			    }
			    if err != nil {
				common.Debug("=================SavePubKeyDataToDb, re-get db fail and save fail ===============","key",kd.Key)
			    } else {
				db = dbtmp
				err = db.Put(kd.Key, []byte(kd.Data))
				if err != nil {
					common.Debug("=================SavePubKeyDataToDb, re-get db success and save fail ===============","key",kd.Key)
				}
			    }

			}
		    }
		} else {
			common.Debug("=================SavePubKeyDataToDb, save to db fail ,db is nil ===============","key",kd.Key)
		}

		time.Sleep(time.Duration(1000000)) //na, 1 s = 10e9 na
	    }
}

func SaveSkU1ToDb() {
	for {
		kd := <-SkU1Chan
		if dbsk != nil {
		    cm,err := EncryptMsg(kd.Data,cur_enode)
		    if err != nil {
			SkU1Chan <- kd
			continue	
		    }

		    err = dbsk.Put(kd.Key, []byte(cm))
		    if err != nil {
			dir := GetSkU1Dir()
			dbsktmp, err := ethdb.NewLDBDatabase(dir, cache, handles)
			//bug
			if err != nil {
				for i := 0; i < 100; i++ {
					dbsktmp, err = ethdb.NewLDBDatabase(dir, cache, handles)
					if err == nil {
						break
					}

					time.Sleep(time.Duration(1000000))
				}
			}
			if err != nil {
			    //dbsk = nil
			} else {
			    dbsk = dbsktmp
			    err = dbsk.Put(kd.Key, []byte(cm))
			    if err != nil {
				SkU1Chan <- kd
				continue
			    }
			}

		    }
		//	db.Close()
		} else {
			SkU1Chan <- kd
		}

		time.Sleep(time.Duration(1000000)) //na, 1 s = 10e9 na
	    }
}

func GetAllPubKeyDataFromDb() *common.SafeMap {
	kd := common.NewSafeMap(10)
	if db != nil {
	    common.Debug("========================GetAllPubKeyDataFromDb,db is not nil =================================")
	    iter := db.NewIterator()
	    for iter.Next() {
		key := string(iter.Key())
		value := string(iter.Value())

		ss, err := UnCompress(value)
		if err == nil {
		    pubs3, err := Decode2(ss, "PubKeyData")
		    if err == nil {
			pd,ok := pubs3.(*PubKeyData)
			if ok {
			    kd.WriteMap(key, pd)
			    continue
			}
		    }
		    
		    pubs, err := Decode2(ss, "AcceptReqAddrData")
		    if err == nil {
			pd,ok := pubs.(*AcceptReqAddrData)
			if ok {
			    kd.WriteMap(key, pd)

				////////ec3////
				/*common.Debug("==================GetAllPubKeyDataFromDb at ec3==============","key",key,"pubkey",pd.PubKey,"initiator",pd.Initiator,"pd.Status",pd.Status)
			       if strings.EqualFold(pd.Initiator,cur_enode) && pd.Status == "Success" {
					go func() {
						PutPreSigal(pd.PubKey,true)

						for {
							if NeedPreSign(pd.PubKey) && GetPreSigal(pd.PubKey) {
								//one,_ := new(big.Int).SetString("1",0)
								//PreSignNonce = new(big.Int).Add(PreSignNonce,one)
								tt := fmt.Sprintf("%v",time.Now().UnixNano()/1e6)
								nonce := Keccak256Hash([]byte(strings.ToLower(pd.PubKey + pd.GroupId + tt))).Hex()
								index := 0
								ids := GetIds2("ECDSA", pd.GroupId)
								for kk, id := range ids {
									enodes := GetEnodesByUid(id, "ECDSA", pd.GroupId)
									if IsCurNode(enodes, cur_enode) {
										if kk >= 2 {
											index = kk - 2
										} else {
											index = kk
										}
										break
									}

								}
								tmp := ids[index:index+3]
								ps := &PreSign{Pub:pd.PubKey,Gid:pd.GroupId,Nonce:nonce,Index:index}

								val,err := Encode2(ps)
								if err != nil {
									common.Debug("=====================GetAllPubKeyDataFromDb at ec3, at start========================","err",err)
									time.Sleep(time.Duration(10000000))
								    continue 
								}
								
								for _, id := range tmp {
									enodes := GetEnodesByUid(id, "ECDSA", pd.GroupId)
									common.Debug("===============GetAllPubKeyDataFromDb at ec3, at start ,get enodes===============","enodes",enodes,"index",index)
									if IsCurNode(enodes, cur_enode) {
										common.Debug("===============GetAllPubKeyDataFromDb at ec3, at start ,get cur enodes===============","enodes",enodes)
										continue
									}
									SendMsgToPeer(enodes, val)
								}

								rch := make(chan interface{}, 1)
								SetUpMsgList3(val,cur_enode,rch)
								_, _,cherr := GetChannelValue(waitall+10,rch)
								if cherr != nil {
									common.Debug("=====================GetAllPubKeyDataFromDb at ec3, at start ========================","cherr",cherr)
								}

								common.Debug("===================generate pre-sign data at start===============","current total number of the data ",GetTotalCount(pd.PubKey),"the number of remaining pre-sign data",(PrePubDataCount-GetTotalCount(pd.PubKey)),"pubkey",pd.PubKey)
							} 

							time.Sleep(time.Duration(1000000))
						}
					}()
			       }*/
				////////////

			    //fmt.Printf("%v ==============GetAllPubKeyDataFromDb,success read AcceptReqAddrData. key = %v,pd = %v ===============\n", common.CurrentTime(), key,pd)
			    continue
			}
		    }
		    
		    /*pubs2, err := Decode2(ss, "AcceptLockOutData")
		    if err == nil {
			pd,ok := pubs2.(*AcceptLockOutData)
			if ok {
			    kd.WriteMap(key, pd)
			    //fmt.Printf("%v ==============GetAllPubKeyDataFromDb,success read AcceptLockOutData. key = %v,pd = %v ===============\n", common.CurrentTime(), key,pd)
			    continue
			}
		    }

		    pubs4, err := Decode2(ss, "AcceptSignData")
		    if err == nil {
			pd,ok := pubs4.(*AcceptSignData)
			if ok {
			    kd.WriteMap(key, pd)
			    //fmt.Printf("%v ==============GetAllPubKeyDataFromDb,success read AcceptReqAddrData. key = %v,pd = %v ===============\n", common.CurrentTime(), key,pd)
			    continue
			}
		    }
		    
		    pubs5, err := Decode2(ss, "AcceptReShareData")
		    if err == nil {
			pd,ok := pubs5.(*AcceptReShareData)
			if ok {
			    kd.WriteMap(key, pd)
			    //fmt.Printf("%v ==============GetAllPubKeyDataFromDb,success read AcceptReqAddrData. key = %v,pd = %v ===============\n", common.CurrentTime(), key,pd)
			    continue
			}
		    }*/
		    
		    continue
		}

		kd.WriteMap(key, []byte(value))
	    }
	    
	    iter.Release()
    //	db.Close()
	}

	return kd
}

func GetValueFromPubKeyData(key string) (bool,interface{}) {
    if key == "" {
	    common.Debug("========================GetValueFromPubKeyData, param err=======================","key",key)
	return false,nil
    }

    datmp, exsit := LdbPubKeyData.ReadMap(key)
    if !exsit {
	    common.Info("========================GetValueFromPubKeyData, get value from memory fail =======================","key",key)
	da := GetPubKeyDataValueFromDb(key)
	if da == nil {
	    common.Info("========================GetValueFromPubKeyData, get value from local db fail =======================","key",key)
	    return false,nil
	}

	ss, err := UnCompress(string(da))
	if err != nil {
	    common.Info("========================GetValueFromPubKeyData, uncompress err=======================","err",err,"key",key)
	    return true,da
	}

	pubs3, err := Decode2(ss, "PubKeyData")
	if err == nil {
	    pd,ok := pubs3.(*PubKeyData)
	    if ok {
		return true,pd
	    }
	}
	
	pubs, err := Decode2(ss, "AcceptReqAddrData")
	if err == nil {
	    pd,ok := pubs.(*AcceptReqAddrData)
	    if ok {
		return true,pd
	    }
	}
	
	pubs2, err := Decode2(ss, "AcceptLockOutData")
	if err == nil {
	    pd,ok := pubs2.(*AcceptLockOutData)
	    if ok {
		return true,pd
	    }
	}

	pubs4, err := Decode2(ss, "AcceptSignData")
	if err == nil {
	    pd,ok := pubs4.(*AcceptSignData)
	    if ok {
		return true,pd
	    }
	}
	
	pubs5, err := Decode2(ss, "AcceptReShareData")
	if err == nil {
	    pd,ok := pubs5.(*AcceptReShareData)
	    if ok {
		return true,pd
	    }
	}
	
	return true,da
    }

    return exsit,datmp
}

func GetPubKeyDataFromLocalDb(key string) (bool,interface{}) {
    if key == "" {
	return false,nil
    }

    da := GetPubKeyDataValueFromDb(key)
    if da == nil {
	common.Debug("========================GetPubKeyDataFromLocalDb, get pubkey data from db fail =======================","key",key)
	return false,nil
    }

    ss, err := UnCompress(string(da))
    if err != nil {
	common.Debug("========================GetPubKeyDataFromLocalDb, uncompress err=======================","err",err,"key",key)
	return false,nil
    }

    pubs, err := Decode2(ss, "PubKeyData")
    if err != nil {
	common.Debug("========================GetPubKeyDataFromLocalDb, decode err=======================","err",err,"key",key)
	return false,nil
    }

    pd,ok := pubs.(*PubKeyData)
    if !ok {
	common.Debug("========================GetPubKeyDataFromLocalDb, it is not pubkey data ========================")
	return false,nil
    }

    return true,pd 
}

func GetGroupDir() string { //TODO
	dir := common.DefaultDataDir()
	dir += "/dcrmdata/dcrmdb" + discover.GetLocalID().String() + "group"
	return dir
}

func GetDbDir() string {
	dir := common.DefaultDataDir()
	dir += "/dcrmdata/dcrmdb" + cur_enode
	return dir
}

func GetSkU1Dir() string {
	dir := common.DefaultDataDir()
	dir += "/dcrmdata/sk" + cur_enode
	return dir
}

func GetAllAccountsDir() string {
	dir := common.DefaultDataDir()
	dir += "/dcrmdata/allaccounts" + cur_enode
	return dir
}

func GetAcceptLockOutDir() string {
	dir := common.DefaultDataDir()
	dir += "/dcrmdata/dcrmdb/acceptlockout" + cur_enode
	return dir
}

func GetAcceptReqAddrDir() string {
	dir := common.DefaultDataDir()
	dir += "/dcrmdata/dcrmdb/acceptreqaddr" + cur_enode
	return dir
}

func GetGAccsDir() string {
	dir := common.DefaultDataDir()
	dir += "/dcrmdata/dcrmdb/gaccs" + cur_enode
	return dir
}

