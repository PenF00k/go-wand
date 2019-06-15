package remgo

import (
	"cliwand/proto_client"
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/sirupsen/logrus"
	"github.com/PenF00k/go-wand/goapi/remgo/generated"
	"google.golang.org/grpc"
	"io"
	"testing"
)

func TestServeWs(t *testing.T) {
	t.Skip("Only for debug")

	opts := []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial("192.168.88.19:9009", opts...)
	if err != nil {
		t.Fatalf("error on Dial: %v", err)
	}
	defer conn.Close()

	client := debugwebsocket.NewDebugClient(conn)

	ctx := context.Background()

	args := proto_client.RequestOTPArgs{
		Phone: "79161500219",
	}

	bytes, err := proto.Marshal(&args)
	if err != nil {
		t.Fatalf("error on Marshal: %v", err)
	}

	argsObj := debugwebsocket.CallMethodArgs{
		MethodName: "RequestOTP",
		Args:       bytes,
	}

	payload, err := client.CallMethod(ctx, &argsObj)
	if err != nil {
		t.Fatalf("error on CallMethod: %v", err)
	}

	res := &wrappers.StringValue{}
	err = proto.Unmarshal(payload.Value, res)
	if err != nil {
		t.Fatalf("error on Unmarshal: %v", err)
	}

	logrus.Printf("result: %+v", res)
}

func TestSubs(t *testing.T) {
	//t.Skip("Only for debug")
	opts := []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial("192.168.88.19:9009", opts...)
	if err != nil {
		t.Fatalf("error on Dial: %v", err)
	}
	defer conn.Close()

	client := debugwebsocket.NewDebugClient(conn)

	ctx := context.Background()

	login(t, client, ctx)

	eventStream, err := client.RegisterEventCallback(ctx, &debugwebsocket.Empty{})
	if err != nil {
		t.Fatalf("error on RegisterEventCallback: %v", err)
	}

	subsArgs := proto_client.SubscribeProjectArgs{
		ProjectID: "9093af80-60bf-4b74-9885-41297af19a7c",
	}

	argsBytes, err := proto.Marshal(&subsArgs)
	if err != nil {
		t.Fatalf("error on Marshal: %v", err)
	}

	args := debugwebsocket.SubscribeArgs{
		Args:             argsBytes,
		SubscriptionName: "SubscribeProject",
	}

	//subsArgs := proto_client.SubscribeOwnedProjectsArgs{}
	//
	//argsBytes, err := proto.Marshal(&subsArgs)
	//if err != nil {
	//	t.Fatalf("error on Marshal: %v", err)
	//}
	//
	//args := debugwebsocket.SubscribeArgs{
	//	Args:             argsBytes,
	//	SubscriptionName: "SubscribeOwnedProjects",
	//}

	_, err = client.Subscribe(ctx, &args)
	if err != nil {
		t.Fatalf("error on Subscribe: %v", err)
	}

	for {
		event, err := eventStream.Recv()
		if err == io.EOF {
			logrus.Warnf("got eof on Recv")
			break
		}
		if err != nil {
			logrus.Errorf("error on Recv: %v", err)
		}
		logrus.Printf("got event %+v: data = %+v", event.FullSubscriptionName, event.Data)
	}

}

func login(t *testing.T, client debugwebsocket.DebugClient, ctx context.Context) {
	phone := "+79161500219"
	args := proto_client.RequestOTPArgs{
		Phone: phone,
	}

	bytes, err := proto.Marshal(&args)
	if err != nil {
		t.Fatalf("error on Marshal: %v", err)
	}

	argsObj := debugwebsocket.CallMethodArgs{
		MethodName: "RequestOTP",
		Args:       bytes,
	}

	payload, err := client.CallMethod(ctx, &argsObj)
	if err != nil {
		t.Fatalf("error on CallMethod: %v", err)
	}

	res := &wrappers.StringValue{}
	err = proto.Unmarshal(payload.Value, res)
	if err != nil {
		t.Fatalf("error on Unmarshal: %v", err)
	}

	logrus.Printf("result: %+v", res)

	passArgs := proto_client.ValidateOTPCustomerArgs{
		Phone:     phone,
		Code:      "1234",
		RequestID: res.Value,
	}

	bytes, err = proto.Marshal(&passArgs)
	if err != nil {
		t.Fatalf("error on Marshal: %v", err)
	}

	argsObj = debugwebsocket.CallMethodArgs{
		MethodName: "ValidateOTPCustomer",
		Args:       bytes,
	}

	payload, err = client.CallMethod(ctx, &argsObj)
	if err != nil {
		t.Fatalf("error on CallMethod: %v", err)
	}

	device := &proto_client.DeviceInfoDto{}
	err = proto.Unmarshal(payload.Value, device)
	if err != nil {
		t.Fatalf("error on Unmarshal: %v", err)
	}

	logrus.Printf("result: %+v", device)
}

func TestIstokenValid(t *testing.T) {
	//t.Skip("Only for debug")
	opts := []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial("192.168.88.19:9009", opts...)
	if err != nil {
		t.Fatalf("error on Dial: %v", err)
	}
	defer conn.Close()

	client := debugwebsocket.NewDebugClient(conn)

	ctx := context.Background()

	args := proto_client.IsTokenValidArgs{}

	bytes, err := proto.Marshal(&args)
	if err != nil {
		t.Fatalf("error on Marshal: %v", err)
	}

	argsObj := debugwebsocket.CallMethodArgs{
		MethodName: "IsTokenValid",
		Args:       bytes,
	}

	payload, err := client.CallMethod(ctx, &argsObj)
	if err != nil {
		t.Fatalf("error on CallMethod: %v", err)
	}

	res := &wrappers.BoolValue{}
	err = proto.Unmarshal(payload.Value, res)
	if err != nil {
		t.Fatalf("error on Unmarshal: %v", err)
	}

	logrus.Printf("result: %+v", res)

}
