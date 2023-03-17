package config

// 系统配置，对应yml
// viper内置了mapstructure, yml文件用"-"区分单词, 转为驼峰方便

var (
	LogOperateConfigStr = `
<seelog type="asynctimer" asyncinterval="1000" minlevel="trace" maxlevel="error">  
	<outputs formatid="common">  
		<buffered formatid="common" size="1048576" flushperiod="1000">  
			<rollingfile type="size" filename="/var/loger/cmdb-notify-operate.log" maxsize="104857600" maxrolls="10"/>  
		</buffered>
	</outputs>  	  
	 <formats>
		 <format id="common" format="%Date %Time [%LEV] [%File:%Line] [%Func] %Msg%n" />  
	 </formats>  
</seelog>
`
	LogAccessConfigStr = `
<seelog type="asynctimer" asyncinterval="1000" minlevel="trace" maxlevel="error">  
	<outputs formatid="common">  
		<buffered formatid="common" size="1048576" flushperiod="1000">  
			<rollingfile type="size" filename="/var/loger/cmdb-notify-access.log" maxsize="104857600" maxrolls="10"/>  
		</buffered>
	</outputs>  	  
	 <formats>
		 <format id="common" format="%Date %Time [%LEV] [%File:%Line] [%Func] %Msg%n" />  
	 </formats>  
</seelog>
`
)

//// 全局配置变量
//var Conf = new(config)
//
//type config struct {
//	System    *SystemConfig    `mapstructure:"system" json:"system"`
//	Logs      *LogsConfig      `mapstructure:"logs" json:"logs"`
//	Mysql     *MysqlConfig     `mapstructure:"mysql" json:"mysql"`
//	Casbin    *CasbinConfig    `mapstructure:"casbin" json:"casbin"`
//	Jwt       *JwtConfig       `mapstructure:"jwt" json:"jwt"`
//	RateLimit *RateLimitConfig `mapstructure:"rate-limit" json:"rateLimit"`
//}
//
//// 设置读取配置信息
//func InitConfig() {
//	workDir, err := os.Getwd()
//	if err != nil {
//		panic(fmt.Errorf("读取应用目录失败:%s", err))
//	}
//	viper.SetConfigName("config")
//	viper.SetConfigType("yml")
//	viper.AddConfigPath(workDir + "/")
//	// 读取配置信息
//	err = viper.ReadInConfig()
//
//	// 热更新配置
//	viper.WatchConfig()
//	viper.OnConfigChange(func(e fsnotify.Event) {
//		// 将读取的配置信息保存至全局变量Conf
//		if err := viper.Unmarshal(Conf); err != nil {
//			panic(fmt.Errorf("初始化配置文件失败:%s", err))
//		}
//		// 读取rsa key
//		Conf.System.RSAPublicBytes = RSAReadKeyFromFile(Conf.System.RSAPublicKey)
//		Conf.System.RSAPrivateBytes = RSAReadKeyFromFile(Conf.System.RSAPrivateKey)
//	})
//
//	if err != nil {
//		panic(fmt.Errorf("读取配置文件失败:%s", err))
//	}
//	// 将读取的配置信息保存至全局变量Conf
//	if err := viper.Unmarshal(Conf); err != nil {
//		panic(fmt.Errorf("初始化配置文件失败:%s", err))
//	}
//	fmt.Println(Conf)
//	// 读取rsa key
//	Conf.System.RSAPublicBytes = RSAReadKeyFromFile(Conf.System.RSAPublicKey)
//	Conf.System.RSAPrivateBytes = RSAReadKeyFromFile(Conf.System.RSAPrivateKey)
//
//}
