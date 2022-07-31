# versionpb

基于 protobuf 自定义选项，实现的数据多版本定义方法

## 例子

```protobuf
syntax = "proto3";
package examplepb;
option go_package = "examplepb/";
import "github.com/fananchong/versionpb/version.proto";

message Msg1 {
  option (versionpb.version_msg) = "3.0";
  enum Enum1 {
    option (versionpb.version_enum) = "3.0";
    E1 = 0;
    E2 = 1;
    E3 = 2 [ (versionpb.version_enum_value) = "3.3" ];
  }
  bytes f1 = 1;
  int64 f2 = 2;
  Enum1 f3 = 3;
  bool f4 = 4;
  int64 f5 = 5 [ (versionpb.version_field) = "3.1" ];
  Enum1 f6 = 6 [ (versionpb.version_field) = "3.2" ];
}
```


完整例子参见： [https://github.com/fananchong/use_protobuf_define_multi_version_example](https://github.com/fananchong/use_protobuf_define_multi_version_example)
