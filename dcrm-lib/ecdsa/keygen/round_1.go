package keygen

import (
	"errors"
	"fmt"
	"math/big"
	"github.com/anyswap/Anyswap-MPCNode/dcrm-lib/crypto/ec2"
	"github.com/anyswap/Anyswap-MPCNode/dcrm-lib/dcrm"
	"github.com/anyswap/Anyswap-MPCNode/crypto/secp256k1"
)

func (round *round1) Start() error {
	if round.started {
	    fmt.Printf("============ round1 start error,already started============\n")
	    return errors.New("round already started")
	}
	round.number = 1
	round.started = true
	round.resetOK()

	u1 := dcrm.GetRandomIntFromZn(secp256k1.S256().N)
	c1 := dcrm.GetRandomIntFromZn(secp256k1.S256().N)
	u1Poly, u1PolyG, _ := ec2.Vss2Init(u1, round.threshold)
	_, c1PolyG, _ := ec2.Vss2Init(c1, round.threshold)
	
	u1Gx, u1Gy := secp256k1.S256().ScalarBaseMult(u1.Bytes())
	u1Secrets := make([]*big.Int, 0)
	u1Secrets = append(u1Secrets, u1Gx)
	u1Secrets = append(u1Secrets, u1Gy)
	for i := 1; i < len(u1PolyG.PolyG); i++ {
		u1Secrets = append(u1Secrets, u1PolyG.PolyG[i][0])
		u1Secrets = append(u1Secrets, u1PolyG.PolyG[i][1])
	}
	commitU1G := new(ec2.Commitment).Commit(u1Secrets...)

	//bip32
	c1Gx, c1Gy := secp256k1.S256().ScalarBaseMult(c1.Bytes())
	c1Secrets := make([]*big.Int, 0)
	c1Secrets = append(c1Secrets, c1Gx)
	c1Secrets = append(c1Secrets, c1Gy)
	for i := 1; i < len(c1PolyG.PolyG); i++ {
		c1Secrets = append(c1Secrets, c1PolyG.PolyG[i][0])
		c1Secrets = append(c1Secrets, c1PolyG.PolyG[i][1])
	}
	commitC1G := new(ec2.Commitment).Commit(c1Secrets...)

	// 3. generate their own paillier public key and private key
	u1PaillierPk, u1PaillierSk := ec2.GenerateKeyPair(round.paillierkeylength)

	round.temp.u1 = u1
	round.temp.u1Poly = u1Poly
	round.temp.u1PolyG = u1PolyG
	round.temp.commitU1G = commitU1G
	round.temp.c1 = c1
	round.temp.commitC1G = commitC1G
	round.temp.u1PaillierPk = u1PaillierPk
	round.temp.u1PaillierSk = u1PaillierSk

	index,err := round.GetDNodeIDIndex(round.dnodeid)
	if err != nil {
	    fmt.Printf("============round1 start,get dnode id index fail,err = %v ===========\n",err)
	    return err
	}

	kg := &dcrm.KGRound1Message{
	    KGRoundMessage:new(dcrm.KGRoundMessage),
	    ComC:commitU1G.C,
	    ComC_bip32:commitC1G.C,
	    U1PaillierPk:u1PaillierPk,
	}
	kg.SetFromID(round.dnodeid)
	kg.SetFromIndex(index)

	round.Save.U1PaillierSk = u1PaillierSk
	round.Save.U1PaillierPk[index] = u1PaillierPk
	round.temp.kgRound1Messages[index] = kg
	round.out <- kg

	fmt.Printf("============ round1 start success ============\n")
	return nil
}

func (round *round1) CanAccept(msg dcrm.Message) bool {
	if _, ok := msg.(*dcrm.KGRound1Message); ok {
		return msg.IsBroadcast()
	}
	return false
}

func (round *round1) Update() (bool, error) {
	for j, msg := range round.temp.kgRound1Messages {
		if round.ok[j] {
			continue
		}
		if msg == nil || !round.CanAccept(msg) {
			return false, nil
		}
		round.ok[j] = true
	}
	
	return true, nil
}

func (round *round1) NextRound() dcrm.Round {
	round.started = false
	return &round2{round}
}

