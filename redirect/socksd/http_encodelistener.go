package socksd

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/ssoor/youniverse/log"
)

const MaxHeaderSize = 4
const MaxBufferSize = 0x1000
const MaxEncodeSize = uint16(0xFFFF)

// NewHTTPLPProxy constructs one HTTPLPProxy
func NewHTTPEncodeListener(l net.Listener) *LPListener {
	return &LPListener{listener: l}
}

type ECipherConn struct {
	net.Conn
	rwc io.ReadWriteCloser

	isPass     bool
	sendHeader []byte

	decodeSize int
	decodeCode byte
	decodeHead [MaxHeaderSize]byte
}

func (this *ECipherConn) getEncodeSize(encodeHeader []byte) (int, error) {
	if encodeHeader[3] != (encodeHeader[0] ^ (encodeHeader[1] + encodeHeader[2])) {
		return 0, errors.New(fmt.Sprint("encryption header information check fails: ", encodeHeader[3], ",Unexpected value: ", (encodeHeader[0] ^ encodeHeader[1] + encodeHeader[2])))
	}

	return int(binary.BigEndian.Uint16(encodeHeader[1:3])), nil
}

func (econn *ECipherConn) Read(data []byte) (lenght int, err error) {
	var decodeSize int

	if 0 != len(econn.sendHeader) { // 发送缓冲区中的数据
		bufSize := len(data)

		if bufSize > len(econn.sendHeader) {
			bufSize = len(econn.sendHeader)
		}

		err = nil
		lenght = copy(data, econn.sendHeader[:bufSize])

		econn.sendHeader = econn.sendHeader[bufSize:]
		//log.Info("Socksd header data is ", string(data[:lenght]))
		return
	}

	if econn.isPass { // 后续数据不用解密 ,直接调用原始函数 - isPass 由 readDecodeHeader() 函数设置
		lenght, err = econn.rwc.Read(data)
		//log.Warning(string(data[:lenght]))
		return
	}

	if 0 == econn.decodeSize { // 当前需要解密的数据为0，准备接受下一个加密头
		lenght, err = econn.readDecodeHeader(data)
	} else { // 解密大小已获得,进入解密流程
		decodeSize, err = econn.rwc.Read(data)
	}

	if 0 != decodeSize {
		lenght = decodeSize // 填充返回数据长度

		if decodeSize > econn.decodeSize {
			decodeSize = econn.decodeSize // 修正解密长度
		}

		for i := 0; i < decodeSize; i++ {
			data[i] ^= econn.decodeCode | 0x80
		}

		econn.decodeSize -= decodeSize
	}

	//log.Info(string(data[:lenght]))
	return
}

func (this *ECipherConn) readDecodeHeader(data []byte) (lenght int, err error) {
	this.isPass = true // 一个新的数据包,默认不需要解密，直接放过

	if lenght, err = io.ReadFull(this.rwc, this.decodeHead[:MaxHeaderSize]); nil != err { // 检测数据包是否为加密包或者有效的 HTTP 包
		if io.ErrUnexpectedEOF == err {
			err = io.EOF
		}

		if io.EOF == err {
			this.isPass = false
		} else {
			log.Warning("Socket full reading failed, current read data:", string(this.decodeHead[:lenght]), "(", lenght, "), need read size:", MaxHeaderSize, " err is:", err)
		}
		return copy(data, this.decodeHead[:lenght]), err
	}

	this.sendHeader = this.decodeHead[:MaxHeaderSize] // 数据需要发送

	if lenght, err = this.getEncodeSize(this.decodeHead[:MaxHeaderSize]); nil == err && lenght <= int(MaxEncodeSize) {
		this.decodeSize = lenght
		this.decodeCode = this.decodeHead[3]

		this.isPass = false // 数据需要解密
		switch this.decodeHead[0] {
		case 0xCD: // GET
			this.sendHeader[0] = 'G'
			this.sendHeader[1] = 'E'
			this.sendHeader[2] = 'T'
			this.sendHeader[3] = ' '
		case 0xDC: // POST
			this.sendHeader[0] = 'P'
			this.sendHeader[1] = 'O'
			this.sendHeader[2] = 'S'
			this.sendHeader[3] = 'T'
		case 0x00: // CONNNECT
			this.sendHeader[0] = 'C'
			this.sendHeader[1] = 'O'
			this.sendHeader[2] = 'N'
			this.sendHeader[3] = 'N'
		case 0xF0: // PUT
			this.sendHeader[0] = 'P'
			this.sendHeader[1] = 'U'
			this.sendHeader[2] = 'T'
			this.sendHeader[3] = ' '
		case 0xF1: // HEAD
			this.sendHeader[0] = 'H'
			this.sendHeader[1] = 'E'
			this.sendHeader[2] = 'A'
			this.sendHeader[3] = 'D'
		case 0xF2: // TRACE
			this.sendHeader[0] = 'T'
			this.sendHeader[1] = 'R'
			this.sendHeader[2] = 'A'
			this.sendHeader[3] = 'C'
		case 0xF3: // DELECT
			this.sendHeader[0] = 'D'
			this.sendHeader[1] = 'E'
			this.sendHeader[2] = 'L'
			this.sendHeader[3] = 'E'
		default:
			log.Warningf("Unknown socksd encode type: % 2x , encode len: %d\n", this.decodeHead[0], this.decodeSize)
		}

		//log.Infof("Socksd encode code: % 5d , encode len: %d\n", this.decodeCode, this.decodeSize)
	} else {
		log.Warning("Socksd decode failed, current encode data is:", this.decodeHead, string(this.decodeHead[:]))
	}

	return 0, nil
}

func (c *ECipherConn) Write(data []byte) (int, error) {
	return c.rwc.Write(data)
}

func (c *ECipherConn) Close() error {
	err := c.Conn.Close()
	c.rwc.Close()
	return err
}

type LPListener struct {
	listener net.Listener
}

func (this *LPListener) Accept() (c net.Conn, err error) {
	conn, err := this.listener.Accept()

	if err != nil {
		return nil, err
	}

	return &ECipherConn{
		Conn: conn,
		rwc:  conn,

		isPass:     false,
		sendHeader: nil,
	}, nil
}

func (this *LPListener) Close() error {
	return this.listener.Close()
}

func (this *LPListener) Addr() net.Addr {
	return this.listener.Addr()
}
