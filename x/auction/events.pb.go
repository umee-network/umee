// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: umee/auction/v1/events.proto

package auction

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// EventRewardsAuctionResult is emitted at the end of each auction that has at least one bidder.
type EventRewardsAuctionResult struct {
	Id     uint32 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Bidder string `protobuf:"bytes,2,opt,name=bidder,proto3" json:"bidder,omitempty"`
	// Auctioned tokens.
	Rewards []types.Coin `protobuf:"bytes,4,rep,name=rewards,proto3" json:"rewards"`
}

func (m *EventRewardsAuctionResult) Reset()         { *m = EventRewardsAuctionResult{} }
func (m *EventRewardsAuctionResult) String() string { return proto.CompactTextString(m) }
func (*EventRewardsAuctionResult) ProtoMessage()    {}
func (*EventRewardsAuctionResult) Descriptor() ([]byte, []int) {
	return fileDescriptor_b5998c755af8caa1, []int{0}
}
func (m *EventRewardsAuctionResult) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EventRewardsAuctionResult) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EventRewardsAuctionResult.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EventRewardsAuctionResult) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EventRewardsAuctionResult.Merge(m, src)
}
func (m *EventRewardsAuctionResult) XXX_Size() int {
	return m.Size()
}
func (m *EventRewardsAuctionResult) XXX_DiscardUnknown() {
	xxx_messageInfo_EventRewardsAuctionResult.DiscardUnknown(m)
}

var xxx_messageInfo_EventRewardsAuctionResult proto.InternalMessageInfo

func init() {
	proto.RegisterType((*EventRewardsAuctionResult)(nil), "umee.auction.v1.EventRewardsAuctionResult")
}

func init() { proto.RegisterFile("umee/auction/v1/events.proto", fileDescriptor_b5998c755af8caa1) }

var fileDescriptor_b5998c755af8caa1 = []byte{
	// 294 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x34, 0x90, 0xb1, 0x4a, 0xc4, 0x30,
	0x18, 0xc7, 0x9b, 0xf3, 0x38, 0xb1, 0xa2, 0x42, 0xb9, 0xa1, 0x3d, 0x24, 0x16, 0xa7, 0x3a, 0x5c,
	0x62, 0x15, 0x04, 0xc7, 0xab, 0x88, 0x7b, 0xdd, 0x5c, 0xa4, 0x6d, 0x42, 0x0d, 0xda, 0x44, 0x92,
	0xb4, 0xe7, 0x63, 0x38, 0xfa, 0x20, 0x3e, 0x44, 0xc7, 0xc3, 0xc9, 0x49, 0xb4, 0x7d, 0x11, 0x69,
	0x93, 0xdb, 0xf2, 0xe7, 0xf7, 0x23, 0xff, 0xef, 0xfb, 0xdc, 0xe3, 0xba, 0xa2, 0x14, 0x67, 0x75,
	0xa1, 0x99, 0xe0, 0xb8, 0x89, 0x31, 0x6d, 0x28, 0xd7, 0x0a, 0xbd, 0x4a, 0xa1, 0x85, 0x77, 0x34,
	0x50, 0x64, 0x29, 0x6a, 0xe2, 0xc5, 0xbc, 0x14, 0xa5, 0x18, 0x19, 0x1e, 0x5e, 0x46, 0x5b, 0x04,
	0x85, 0x50, 0x95, 0x50, 0x8f, 0x06, 0x98, 0x60, 0x11, 0x34, 0x09, 0xe7, 0x99, 0xa2, 0xb8, 0x89,
	0x73, 0xaa, 0xb3, 0x18, 0x17, 0x82, 0x71, 0xc3, 0x4f, 0x3f, 0x80, 0x1b, 0xdc, 0x0e, 0x95, 0x29,
	0x5d, 0x67, 0x92, 0xa8, 0x95, 0xe9, 0x4a, 0xa9, 0xaa, 0x5f, 0xb4, 0x77, 0xe8, 0x4e, 0x18, 0xf1,
	0x41, 0x08, 0xa2, 0x83, 0x74, 0xc2, 0x88, 0x77, 0xee, 0xce, 0x72, 0x46, 0x08, 0x95, 0xfe, 0x24,
	0x04, 0xd1, 0x5e, 0xe2, 0x7f, 0x7d, 0x2e, 0xe7, 0xb6, 0x6f, 0x45, 0x88, 0xa4, 0x4a, 0xdd, 0x6b,
	0xc9, 0x78, 0x99, 0x5a, 0xcf, 0xbb, 0x76, 0x77, 0xa5, 0xf9, 0xd9, 0x9f, 0x86, 0x3b, 0xd1, 0xfe,
	0x45, 0x80, 0xac, 0x3f, 0x4c, 0x84, 0xec, 0x44, 0xe8, 0x46, 0x30, 0x9e, 0x4c, 0xdb, 0x9f, 0x13,
	0x27, 0xdd, 0xfa, 0xc9, 0x5d, 0xfb, 0x07, 0x9d, 0xb6, 0x83, 0x60, 0xd3, 0x41, 0xf0, 0xdb, 0x41,
	0xf0, 0xde, 0x43, 0x67, 0xd3, 0x43, 0xe7, 0xbb, 0x87, 0xce, 0xc3, 0x59, 0xc9, 0xf4, 0x53, 0x9d,
	0xa3, 0x42, 0x54, 0x78, 0xb8, 0xd2, 0x92, 0x53, 0xbd, 0x16, 0xf2, 0x79, 0x0c, 0xb8, 0xb9, 0xc2,
	0x6f, 0xdb, 0xab, 0xe6, 0xb3, 0x71, 0xd5, 0xcb, 0xff, 0x00, 0x00, 0x00, 0xff, 0xff, 0x05, 0x11,
	0xf0, 0x01, 0x6c, 0x01, 0x00, 0x00,
}

func (m *EventRewardsAuctionResult) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EventRewardsAuctionResult) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EventRewardsAuctionResult) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Rewards) > 0 {
		for iNdEx := len(m.Rewards) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Rewards[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintEvents(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x22
		}
	}
	if len(m.Bidder) > 0 {
		i -= len(m.Bidder)
		copy(dAtA[i:], m.Bidder)
		i = encodeVarintEvents(dAtA, i, uint64(len(m.Bidder)))
		i--
		dAtA[i] = 0x12
	}
	if m.Id != 0 {
		i = encodeVarintEvents(dAtA, i, uint64(m.Id))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintEvents(dAtA []byte, offset int, v uint64) int {
	offset -= sovEvents(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *EventRewardsAuctionResult) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Id != 0 {
		n += 1 + sovEvents(uint64(m.Id))
	}
	l = len(m.Bidder)
	if l > 0 {
		n += 1 + l + sovEvents(uint64(l))
	}
	if len(m.Rewards) > 0 {
		for _, e := range m.Rewards {
			l = e.Size()
			n += 1 + l + sovEvents(uint64(l))
		}
	}
	return n
}

func sovEvents(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozEvents(x uint64) (n int) {
	return sovEvents(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *EventRewardsAuctionResult) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowEvents
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: EventRewardsAuctionResult: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EventRewardsAuctionResult: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			m.Id = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Id |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Bidder", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthEvents
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthEvents
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Bidder = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Rewards", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthEvents
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthEvents
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Rewards = append(m.Rewards, types.Coin{})
			if err := m.Rewards[len(m.Rewards)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipEvents(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthEvents
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipEvents(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowEvents
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowEvents
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowEvents
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthEvents
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupEvents
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthEvents
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthEvents        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowEvents          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupEvents = fmt.Errorf("proto: unexpected end of group")
)
