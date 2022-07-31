package versionpb

import (
	"fmt"
	"strings"

	"github.com/coreos/go-semver/semver"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func MinimalVersion(msg proto.Message) *semver.Version {
	var maxVer *semver.Version
	err := VisitMsg(msg, func(path protoreflect.FullName, ver *semver.Version) error {
		maxVer = maxVersion(maxVer, ver)
		return nil
	})
	if err != nil {
		panic(err)
	}
	return maxVer
}

type Visitor func(path protoreflect.FullName, ver *semver.Version) error

// VisitMsg 根据消息，获取消息的版本
func VisitMsg(msg proto.Message, visitor Visitor) error {
	return visitMessage(msg.ProtoReflect(), visitor)
}

// VisitFileDescriptor 根据 proto 文件描述，获取消息的版本
func VisitFileDescriptor(file protoreflect.FileDescriptor, visitor Visitor) error {
	msgs := file.Messages()
	for i := 0; i < msgs.Len(); i++ {
		err := visitMessageDescriptor(msgs.Get(i), visitor)
		if err != nil {
			return err
		}
	}
	enums := file.Enums()
	for i := 0; i < enums.Len(); i++ {
		err := visitEnumDescriptor(enums.Get(i), visitor)
		if err != nil {
			return err
		}
	}
	return nil
}

func visitMessageDescriptor(md protoreflect.MessageDescriptor, visitor Visitor) error {
	err := visitDescriptor(md, visitor)
	if err != nil {
		return err
	}
	fields := md.Fields()
	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		err = visitDescriptor(fd, visitor)
		if err != nil {
			return err
		}
	}

	enums := md.Enums()
	for i := 0; i < enums.Len(); i++ {
		err := visitEnumDescriptor(enums.Get(i), visitor)
		if err != nil {
			return err
		}
	}
	return err
}

func visitMessage(m protoreflect.Message, visitor Visitor) error {
	md := m.Descriptor()
	err := visitDescriptor(md, visitor)
	if err != nil {
		return err
	}
	m.Range(func(field protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		fd := md.Fields().Get(field.Index())
		err = visitDescriptor(fd, visitor)
		if err != nil {
			return false
		}

		switch m := value.Interface().(type) {
		case protoreflect.Message:
			err = visitMessage(m, visitor)
		case protoreflect.EnumNumber:
			err = visitEnumNumber(fd.Enum(), m, visitor)
		}
		if err != nil {
			return false
		}
		return true
	})
	return err
}

func visitEnumDescriptor(enum protoreflect.EnumDescriptor, visitor Visitor) error {
	err := visitDescriptor(enum, visitor)
	if err != nil {
		return err
	}
	fields := enum.Values()
	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		err = visitDescriptor(fd, visitor)
		if err != nil {
			return err
		}
	}
	return err
}

func visitEnumNumber(enum protoreflect.EnumDescriptor, number protoreflect.EnumNumber, visitor Visitor) error {
	err := visitDescriptor(enum, visitor)
	if err != nil {
		return err
	}
	intNumber := int(number)
	fields := enum.Values()
	if intNumber >= fields.Len() || intNumber < 0 {
		return fmt.Errorf("could not visit EnumNumber [%d]", intNumber)
	}
	return visitEnumValue(fields.Get(intNumber), visitor)
}

func visitEnumValue(enum protoreflect.EnumValueDescriptor, visitor Visitor) error {
	valueOpts := enum.Options().(*descriptorpb.EnumValueOptions)
	if valueOpts != nil {
		ver, _ := versionFromOptionsString(valueOpts.String())
		err := visitor(enum.FullName(), ver)
		if err != nil {
			return err
		}
	}
	return nil
}

func visitDescriptor(md protoreflect.Descriptor, visitor Visitor) error {
	opts, ok := md.Options().(fmt.Stringer)
	if !ok {
		return nil
	}
	ver, err := versionFromOptionsString(opts.String())
	if err != nil {
		return fmt.Errorf("%s: %s", md.FullName(), err)
	}
	return visitor(md.FullName(), ver)
}

func maxVersion(a *semver.Version, b *semver.Version) *semver.Version {
	if a != nil && (b == nil || b.LessThan(*a)) {
		return a
	}
	return b
}

func versionFromOptionsString(opts string) (*semver.Version, error) {
	msgs := []string{"[version_msg]:", "[version_field]:", "[version_enum]:", "[version_enum_value]:"}
	var end, index int
	for _, msg := range msgs {
		index = strings.Index(opts, msg)
		end = index + len(msg)
		if index != -1 {
			break
		}
	}
	if index == -1 {
		return nil, nil
	}
	var verStr string
	_, err := fmt.Sscanf(opts[end:], "%q", &verStr)
	if err != nil {
		return nil, err
	}
	if strings.Count(verStr, ".") == 1 {
		verStr = verStr + ".0"
	}
	ver, err := semver.NewVersion(verStr)
	if err != nil {
		return nil, err
	}
	return ver, nil
}
