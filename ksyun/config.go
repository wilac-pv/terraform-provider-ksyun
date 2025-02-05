package ksyun

import (
	"fmt"
	"github.com/KscSDK/ksc-sdk-go/ksc"
	"github.com/KscSDK/ksc-sdk-go/ksc/utils"
	"github.com/KscSDK/ksc-sdk-go/service/bws"
	"github.com/KscSDK/ksc-sdk-go/service/ebs"
	"github.com/KscSDK/ksc-sdk-go/service/eip"
	"github.com/KscSDK/ksc-sdk-go/service/epc"
	"github.com/KscSDK/ksc-sdk-go/service/iam"
	"github.com/KscSDK/ksc-sdk-go/service/kcm"
	"github.com/KscSDK/ksc-sdk-go/service/kcsv1"
	"github.com/KscSDK/ksc-sdk-go/service/kcsv2"
	"github.com/KscSDK/ksc-sdk-go/service/kec"
	"github.com/KscSDK/ksc-sdk-go/service/krds"
	"github.com/KscSDK/ksc-sdk-go/service/mongodb"
	"github.com/KscSDK/ksc-sdk-go/service/rabbitmq"
	"github.com/KscSDK/ksc-sdk-go/service/sks"
	"github.com/KscSDK/ksc-sdk-go/service/slb"
	"github.com/KscSDK/ksc-sdk-go/service/sqlserver"
	"github.com/KscSDK/ksc-sdk-go/service/tag"
	"github.com/KscSDK/ksc-sdk-go/service/tagv2"
	"github.com/KscSDK/ksc-sdk-go/service/vpc"
	"github.com/wilac-pv/ksyun-ks3-go-sdk/ks3"
	"sync"
)

// Config is the configuration of ksyun meta data
type Config struct {
	AccessKey     string
	SecretKey     string
	Region        string
	Insecure      bool
	Domain        string
	DryRun        bool
	IgnoreService bool
}

// Client will returns a client with connections for all product
func (c *Config) Client() (*KsyunClient, error) {
	var client KsyunClient
	//init ksc client info
	client.region = c.Region
	cli := ksc.NewClient(c.AccessKey, c.SecretKey)
	cfg := &ksc.Config{
		Region: &c.Region,
	}
	client.config = c
	url := &utils.UrlInfo{
		UseSSL:                      false,
		Locate:                      false,
		CustomerDomain:              c.Domain,
		CustomerDomainIgnoreService: c.IgnoreService,
	}
	client.dryRun = c.DryRun
	client.vpcconn = vpc.SdkNew(cli, cfg, url)
	client.eipconn = eip.SdkNew(cli, cfg, url)
	client.slbconn = slb.SdkNew(cli, cfg, url)
	client.kecconn = kec.SdkNew(cli, cfg, url)
	client.sqlserverconn = sqlserver.SdkNew(cli, cfg, url)
	client.krdsconn = krds.SdkNew(cli, cfg, url)
	client.kcmconn = kcm.SdkNew(cli, cfg, url)
	client.sksconn = sks.SdkNew(cli, cfg, url)
	client.kcsv1conn = kcsv1.SdkNew(cli, cfg, url)
	client.kcsv2conn = kcsv2.SdkNew(cli, cfg, url)
	client.epcconn = epc.SdkNew(cli, cfg, url)
	client.ebsconn = ebs.SdkNew(cli, cfg, url)
	client.mongodbconn = mongodb.SdkNew(cli, cfg, url)
	client.iamconn = iam.SdkNew(cli, cfg, url)
	client.rabbitmqconn = rabbitmq.SdkNew(cli, cfg, url)
	client.bwsconn = bws.SdkNew(cli, cfg, url)
	client.tagconn = tagv2.SdkNew(cli, cfg, url)
	client.tagv1conn = tag.SdkNew(cli, cfg, url)

	//懒加载ks3-client 所以不在此初始化
	return &client, nil
}

var goSdkMutex = sync.RWMutex{} // The Go SDK is not thread-safe
var loadSdkfromRemoteMutex = sync.Mutex{}
var loadSdkEndpointMutex = sync.Mutex{}

func (client *KsyunClient) WithKs3BucketByName(bucketName string, do func(*ks3.Bucket) (interface{}, error)) (interface{}, error) {
	return client.WithKs3Client(func(ks3Client *ks3.Client) (interface{}, error) {
		bucket, err := client.ks3conn.Bucket(bucketName)
		if err != nil {
			return nil, fmt.Errorf("unable to get the bucket %s: %#v", bucketName, err)
		}
		return do(bucket)
	})
}
func (client *KsyunClient) WithKs3Client(do func(*ks3.Client) (interface{}, error)) (interface{}, error) {
	goSdkMutex.Lock()
	defer goSdkMutex.Unlock()
	// Initialize the KS3 client if necessary
	if client.ks3conn == nil {
		ks3conn, err := ks3.New(client.config.Domain, client.config.AccessKey, client.config.SecretKey)
		if err != nil {
			return nil, fmt.Errorf("unable to initialize the KS3 client: %#v", err)
		}
		client.ks3conn = ks3conn
	}
	return do(client.ks3conn)
}
