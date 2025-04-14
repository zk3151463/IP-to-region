package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"github.com/oschwald/geoip2-golang"
)

// IPLocation 表示 IP 地理位置信息
type IPLocation struct {
	IP          string  `json:"ip"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Province    string  `json:"province,omitempty"`    // 省份/州
	City        string  `json:"city,omitempty"`        // 城市
	District    string  `json:"district,omitempty"`    // 区县
	ISP         string  `json:"isp,omitempty"`         // 运营商
	Latitude    float64 `json:"latitude,omitempty"`    // 纬度
	Longitude   float64 `json:"longitude,omitempty"`   // 经度
	TimeZone    string  `json:"timezone,omitempty"`    // 时区
	ContinentEN string  `json:"continent_en"`          // 大洲（英文）
	ContinentZH string  `json:"continent_zh"`          // 大洲（中文）
	CountryEN   string  `json:"country_en"`            // 国家（英文）
	CountryZH   string  `json:"country_zh"`            // 国家（中文）
	RegionEN    string  `json:"region_en,omitempty"`   // 地区（英文）
	RegionZH    string  `json:"region_zh,omitempty"`   // 地区（中文）
	CityEN      string  `json:"city_en,omitempty"`     // 城市（英文）
	CityZH      string  `json:"city_zh,omitempty"`     // 城市（中文）
	AccuracyKM  int     `json:"accuracy_km,omitempty"` // 精确度半径(公里)
	Source      string  `json:"source"`                // 数据来源
}

// ErrorResponse 表示错误响应
type ErrorResponse struct {
	Error string `json:"error"`
}

// IPService 接口定义了 IP 查询服务的方法
type IPService interface {
	GetLocation(ip net.IP) (*IPLocation, error)
	Close()
}

// GeoIPService 实现了基于 GeoIP 的 IP 查询服务
type GeoIPService struct {
	db *geoip2.Reader
}

// IP2RegionService 实现了基于 ip2region 的 IP 查询服务
type IP2RegionService struct {
	db *xdb.Searcher
}

// NewGeoIPService 创建一个新的 GeoIP 服务
func NewGeoIPService(dbPath string) (*GeoIPService, error) {
	db, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, err
	}
	return &GeoIPService{db: db}, nil
}

// NewIP2RegionService 创建一个新的 ip2region 服务
func NewIP2RegionService(dbPath string) (*IP2RegionService, error) {
	// 检查文件是否存在
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("ip2region database file not found: %s", dbPath)
	}

	// 加载 xdb 文件
	cBuff, err := xdb.LoadContentFromFile(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load ip2region database: %v", err)
	}

	// 创建查询对象
	searcher, err := xdb.NewWithBuffer(cBuff)
	if err != nil {
		return nil, fmt.Errorf("failed to create ip2region searcher: %v", err)
	}

	return &IP2RegionService{db: searcher}, nil
}

// Close 关闭 GeoIP 服务
func (s *GeoIPService) Close() {
	if s.db != nil {
		s.db.Close()
	}
}

// Close 关闭 ip2region 服务
func (s *IP2RegionService) Close() {
	if s.db != nil {
		s.db.Close()
	}
}

// GetLocation 获取 IP 地理位置信息 (GeoIP)
func (s *GeoIPService) GetLocation(ip net.IP) (*IPLocation, error) {
	record, err := s.db.City(ip)
	if err != nil {
		return nil, err
	}

	// 获取大洲信息
	continent := record.Continent.Names["en"]
	continentZH := record.Continent.Names["zh-CN"]
	if continentZH == "" {
		continentZH = continent // 如果没有中文名称，使用英文
	}

	// 获取国家信息
	country := record.Country.Names["en"]
	countryZH := record.Country.Names["zh-CN"]
	if countryZH == "" {
		countryZH = country
	}

	// 获取地区/省份信息
	region := ""
	regionZH := ""
	if len(record.Subdivisions) > 0 {
		region = record.Subdivisions[0].Names["en"]
		regionZH = record.Subdivisions[0].Names["zh-CN"]
		if regionZH == "" {
			regionZH = region
		}
	}

	// 获取城市信息
	city := record.City.Names["en"]
	cityZH := record.City.Names["zh-CN"]
	if cityZH == "" {
		cityZH = city
	}

	return &IPLocation{
		IP:          ip.String(),
		Country:     countryZH,
		CountryCode: record.Country.IsoCode,
		Province:    regionZH,
		City:        cityZH,
		Latitude:    record.Location.Latitude,
		Longitude:   record.Location.Longitude,
		TimeZone:    record.Location.TimeZone,
		ContinentEN: continent,
		ContinentZH: continentZH,
		CountryEN:   country,
		CountryZH:   countryZH,
		RegionEN:    region,
		RegionZH:    regionZH,
		CityEN:      city,
		CityZH:      cityZH,
		AccuracyKM:  int(record.Location.AccuracyRadius),
		Source:      "GeoIP",
	}, nil
}

// GetLocation 获取 IP 地理位置信息 (ip2region)
func (s *IP2RegionService) GetLocation(ip net.IP) (*IPLocation, error) {
	// 查询 IP 信息
	region, err := s.db.SearchByStr(ip.String())
	if err != nil {
		return nil, err
	}

	// 解析返回的地区信息
	// ip2region 返回格式: 国家|区域|省份|城市|ISP
	parts := strings.Split(region, "|")
	if len(parts) < 5 {
		return nil, fmt.Errorf("invalid region format: %s", region)
	}

	country := parts[0]
	regionName := parts[1]
	province := parts[2]
	city := parts[3]
	isp := parts[4]

	// 处理未知值
	if country == "0" {
		country = "未知"
	}
	if regionName == "0" {
		regionName = ""
	}
	if province == "0" {
		province = ""
	}
	if city == "0" {
		city = ""
	}
	if isp == "0" {
		isp = ""
	}

	// 由于 ip2region 不提供经纬度信息，这里设置为 0
	return &IPLocation{
		IP:          ip.String(),
		Country:     country,
		CountryCode: "", // ip2region 不提供国家代码
		Province:    province,
		City:        city,
		District:    regionName,
		ISP:         isp,
		Latitude:    0,
		Longitude:   0,
		TimeZone:    "",
		ContinentEN: "",
		ContinentZH: "",
		CountryEN:   country,
		CountryZH:   country,
		RegionEN:    regionName,
		RegionZH:    regionName,
		CityEN:      city,
		CityZH:      city,
		AccuracyKM:  0,
		Source:      "ip2region",
	}, nil
}

// getClientIP 获取客户端真实 IP
func getClientIP(c *gin.Context) string {
	// 检查各种可能的请求头
	for _, header := range []string{
		"X-Real-IP",
		"X-Forwarded-For",
		"X-Client-IP",
		"CF-Connecting-IP", // Cloudflare
	} {
		ip := c.GetHeader(header)
		if ip != "" {
			// X-Forwarded-For 可能包含多个 IP，取第一个
			if header == "X-Forwarded-For" {
				ips := strings.Split(ip, ",")
				if len(ips) > 0 {
					return strings.TrimSpace(ips[0])
				}
			}
			return ip
		}
	}

	// 如果没有代理头，则使用远程地址
	return c.ClientIP()
}

func main() {
	// 解析命令行参数
	port := flag.Int("port", 8081, "服务监听端口")
	geoipDBPath := flag.String("geoip-db", "data/GeoLite2-City.mmdb", "GeoIP 数据库文件路径")
	ip2regionDBPath := flag.String("ip2region-db", "data/ip2region/ip2region.xdb", "ip2region 数据库文件路径")
	dbType := flag.String("db-type", "geoip", "使用的数据库类型: geoip 或 ip2region")
	flag.Parse()

	var ipService IPService
	var err error

	// 根据指定的数据库类型创建相应的服务
	switch *dbType {
	case "geoip":
		ipService, err = NewGeoIPService(*geoipDBPath)
		if err != nil {
			log.Fatal("Failed to load GeoIP database:", err)
		}
	case "ip2region":
		ipService, err = NewIP2RegionService(*ip2regionDBPath)
		if err != nil {
			log.Fatal("Failed to load ip2region database:", err)
		}
	default:
		log.Fatalf("Unsupported database type: %s. Use 'geoip' or 'ip2region'", *dbType)
	}
	defer ipService.Close()

	// 创建 Gin 引擎
	r := gin.Default()

	// 配置 CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// IP 查询接口
	r.GET("/api/ip-location", func(c *gin.Context) {
		var ipStr string

		// 获取查询参数中的 IP
		if queryIP := c.Query("ip"); queryIP != "" {
			ipStr = queryIP
		} else {
			// 如果没有提供 IP，则使用请求者的 IP
			ipStr = getClientIP(c)
		}

		// 解析 IP 地址
		ip := net.ParseIP(ipStr)
		if ip == nil {
			c.JSON(400, ErrorResponse{Error: "Invalid IP address"})
			return
		}

		// 查询地理位置信息
		location, err := ipService.GetLocation(ip)
		if err != nil {
			c.JSON(500, ErrorResponse{Error: err.Error()})
			return
		}

		c.JSON(200, location)
	})

	// 健康检查接口
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"db_type": *dbType,
		})
	})

	// 启动服务器
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting server on %s using %s database...\n", addr, *dbType)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
