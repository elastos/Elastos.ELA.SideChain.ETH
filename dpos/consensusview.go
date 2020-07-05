package dpos

import (
	"bytes"
	"time"

	"github.com/elastos/Elastos.ELA/common"
	"github.com/elastos/Elastos.ELA/common/log"
)

type ViewListener interface {
	OnViewChanged(isOnDuty bool, force bool)
}

const (
	consensusReady = iota
	consensusRunning
)

type ConsensusView struct {
	consensusStatus uint32
	viewOffset      uint32
	publicKey       []byte
	signTolerance   time.Duration
	viewStartTime   time.Time
	isDposOnDuty    bool
	producers       *Producers

	listener ViewListener
}

func (v *ConsensusView) resetViewOffset() {
	v.viewOffset = 0
}

func (v *ConsensusView) SetRunning() {
	v.consensusStatus = consensusRunning
}

func (v *ConsensusView) SetReady() {
	v.consensusStatus = consensusReady
}

func (v *ConsensusView) IsRunning() bool {
	return v.consensusStatus == consensusRunning
}

func (v *ConsensusView) IsReady() bool {
	return v.consensusStatus == consensusReady
}

func (v *ConsensusView) TryChangeView(now time.Time) {
	if v.IsRunning() && now.After(v.viewStartTime.Add(v.signTolerance)) {
		Info("[TryChangeView] succeed")
		v.ChangeView(now, false)
	}
}

func (v *ConsensusView) GetProducers() [][]byte {
	return v.producers.GetProducers()
}

func (v *ConsensusView) calculateOffsetTime(startTime time.Time,
	now time.Time) (uint32, time.Duration) {
	duration := now.Sub(startTime)
	offset := duration / v.signTolerance
	offsetTime := duration % v.signTolerance

	return uint32(offset), offsetTime
}

func (v *ConsensusView) ChangeView(now time.Time, force bool) {
	offset, offsetTime := v.calculateOffsetTime(v.viewStartTime, now)
	v.viewStartTime = now.Add(-offsetTime)
	if force {
		offset = 1
		v.ResetView(now)
	}
	v.viewOffset += offset

	if offset > 0 {
		Info("\n\n\n--------------------Change View---------------------")
		Info("viewStartTime:", v.viewStartTime, "nowTime:", now, "offset:", offset, "offsetTime:", offsetTime, "force:", force)
		currentProducer := v.producers.GetNextOnDutyProducer(v.viewOffset)
		v.isDposOnDuty = bytes.Equal(currentProducer, v.publicKey)
		if bytes.Equal(currentProducer, v.producers.producers[0]) {
			v.resetViewOffset()//Prevent viewOffset overflow
		}
		v.DumpInfo()
		Info("\n\n\n")
		if v.listener != nil {
			v.listener.OnViewChanged(v.isDposOnDuty, force)
		}
	}
}

func (v *ConsensusView) DumpInfo() {
	str := "\n"
	for _, signer := range v.producers.producers {
		if v.ProducerIsOnDuty(signer) {
			duty := log.Color(log.Green, common.BytesToHexString(signer)+" onDuty \n")
			str = str + duty
		} else {
			str = str + common.BytesToHexString(signer) + " not onDuty \n"
		}
	}
	Info(str)
}

func (v *ConsensusView) GetViewInterval() time.Duration {
	return v.signTolerance
}

func (v *ConsensusView) GetViewStartTime() time.Time {
	return v.viewStartTime
}

func (v *ConsensusView) GetViewOffset() uint32 {
	return v.viewOffset
}

func (v *ConsensusView) ResetView(t time.Time) {
	v.viewStartTime = t
}

func (v *ConsensusView) IsProducers(account []byte) bool {
	return v.producers.IsProducers(account)
}

func (v *ConsensusView) IsOnduty() bool {
	return v.isDposOnDuty
}

func (v *ConsensusView) ProducerIsOnDuty(account []byte) bool {
	producer := v.producers.GetNextOnDutyProducer(v.viewOffset)
	return bytes.Equal(producer, account)
}

func (v *ConsensusView) IsMajorityAgree(count int) bool {
	return v.producers.IsMajorityAgree(count)
}

func (v *ConsensusView) IsMajorityRejected(count int) bool {
	return v.producers.IsMajorityRejected(count)
}

func (v *ConsensusView) HasArbitersMinorityCount(count int) bool {
	return v.producers.HasArbitersMinorityCount(count)
}

func (v *ConsensusView) HasProducerMajorityCount(count int) bool {
	return v.producers.HasProducerMajorityCount(count)
}

func NewConsensusView(tolerance time.Duration, account []byte,
	producers *Producers, viewListener ViewListener) *ConsensusView {
	c := &ConsensusView{
		consensusStatus: consensusReady,
		viewStartTime:   time.Unix(0, 0),
		viewOffset:      0,
		publicKey:       account,
		signTolerance:   tolerance,
		producers:       producers,
		listener:        viewListener,
	}
	c.isDposOnDuty = c.ProducerIsOnDuty(account)

	return c
}