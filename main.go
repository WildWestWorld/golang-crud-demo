package main

//导入gin
import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"strconv"
	"time"
)

func main() {

	//连接数据库
	dsn := "root:123456@tcp(127.0.0.1:3306)/go-crud-test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		//设置命名逻辑 不设置SingularTable 被迁移的表会默认待上复数
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // use singular table name, table for `User` would be `user` with this option enabled
		},
	})

	fmt.Println(db)
	fmt.Println(err)

	sqlDB, errDB := db.DB()
	//设置连接池（照着设置就行了）
	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(10 * time.Second) //10秒钟

	//设置结构体 此处创建的结构体 用于创建新的表 此处也可以用 gorm.Model代替
	//GORM 依赖
	//`gorm:"primaryKey"` 设置为主键
	//`gorm:"type:varchar(20);not null" 设置在sql内的类型为varchar(20) 不能为空

	//GIN依赖
	//json:"name" json中的名称
	//form:"name" form中的名称 用于前端 不同参数传入

	//binding:"required" 在payload 中是否是 必须传入的
	type List struct {
		ID uint `gorm:"primaryKey"`

		Name    string `gorm:"type:varchar(20);not null" form:"name" json:"name" binding:"required"`
		Address string `gorm:"type:varchar(20);not null" form:"address" json:"address" binding:"required"`
		State   string `gorm:"type:varchar(20);not null" form:"state" json:"state" binding:"required"`
		Phone   string `gorm:"type:varchar(40);not null" form:"phone" json:"phone" binding:"required"`
		Email   string `gorm:"type:varchar(200);not null" form:"email" json:"email" binding:"required"`

		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt gorm.DeletedAt `gorm:"index"`
	}
	// gorm 给我们的迁移功能，我们能用这个功能+ 上面定义的结构体创建数据库的表
	db.AutoMigrate(&List{})

	fmt.Println(errDB)

	//创建接口
	ginAPI := gin.Default()

	//GET
	ginAPI.GET("/", func(context *gin.Context) {
		context.JSON(200, gin.H{
			"result":  "0",
			"message": "请求成功",
		})
	})

	//带条件搜索的GET
	// 举例:匹配的url格式:  /welcome?firstname=Jane&lastname=Doe
	ginAPI.GET("/user", func(context *gin.Context) {
		var data []List
		//fmt.Println(context)
		//设置一个默认的值，若是我们自己的写一些值会替换掉默认的值
		//firstname := context.DefaultQuery("firstname", "Guest")

		//查询当前的网络路径下的name
		name := context.Query("name") // 是 c.Request.URL.Query().Get("name") 的简写

		//nameAfterFormat, _ := url.QueryUnescape(name)

		db.Where("name = %?%", name).Find(&data)

		if len(data) != 0 {
			context.JSON(200, gin.H{
				"result": '0',
				"data":   data,
				"msg":    "查询成功",
			})
		} else {
			context.JSON(200, gin.H{
				"result": '1',
				"msg":    "没有查询到数据",
			})
		}

		context.JSON(200, gin.H{})
	})

	//全部查询
	ginAPI.GET("/user/all", func(context *gin.Context) {
		var dataList []List
		//1.查询全部数据，查询分页数据
		pageNum := context.DefaultQuery("pageNum", "1")
		pageSize := context.DefaultQuery("pageSize", "10")

		//strconv.Atoi 字符串类型 转 数字类型
		//pageNumInt, err := strconv.Atoi(pageNum)
		//if err != nil {
		//	fmt.Println(err)
		//}
		//pageSizeInt, _ := strconv.Atoi(pageNum)
		//startNum := (pageNumInt - 1) * pageSizeInt

		//返回一个总数
		var totalNum int64
		//查询数据库
		//offset相当起始位置    limit相当于起始位置开始的的数量
		//举例 offset 1 limit 2  =>  从1开始的 2个 数据
		// 若是不加Model就无法 查询到数据
		// 获取所有记录
		db.Model(dataList).Count(&totalNum).Limit(-1).Offset(-1).Find(&dataList)
		if len(dataList) != 0 {
			context.JSON(200, gin.H{
				"result": "0",
				"msg":    "查询成功",
				"data": gin.H{
					"list":     dataList,
					"totalNum": totalNum,
					"pageNum":  pageNum,
					"pageSize": pageSize,
				},
			})
		} else {
			context.JSON(200, gin.H{
				"result": "1",
				"msg":    "没有查询到数据",
				//"data":   dataList,
			})
		}
	})

	//分页查询
	ginAPI.GET("/user/list", func(context *gin.Context) {
		var dataList []List
		//1.查询全部数据，查询分页数据
		pageNum := context.DefaultQuery("pageNum", "1")
		pageSize := context.DefaultQuery("pageSize", "10")

		//strconv.Atoi 字符串类型 转 数字类型
		pageNumInt, err := strconv.Atoi(pageNum)
		if err != nil {
			fmt.Println(err)
		}
		pageSizeInt, _ := strconv.Atoi(pageSize)
		startNum := (pageNumInt - 1) * pageSizeInt

		//返回一个总数
		var totalNum int64
		//查询数据库
		//offset相当起始位置    limit相当于起始位置开始的的数量
		//举例 offset 1 limit 2  =>  从1开始的 2个 数据
		// 若是不加Model就无法 查询到数据
		// 获取所有记录
		db.Model(dataList).Count(&totalNum).Limit(pageSizeInt).Offset(startNum).Find(&dataList)
		if len(dataList) != 0 {
			context.JSON(200, gin.H{
				"result": "0",
				"msg":    "查询成功",
				"data": gin.H{
					"list":     dataList,
					"totalNum": totalNum,
					"pageNum":  pageNum,
					"pageSize": pageSize,
				},
			})
		} else {
			context.JSON(200, gin.H{
				"result": "1",
				"msg":    "没有查询到数据",
				//"data":   dataList,
			})
		}
	})

	//POST
	ginAPI.POST("/user/add", func(context *gin.Context) {
		var payload List

		//ShouldBindJSON 判断我们传入的data是否符合接口设定的payload的规格
		err := context.ShouldBindJSON(&payload)
		fmt.Println(err)
		//如果没有错误 也就是添加成功 我们进行数据库操作
		if err == nil {
			context.JSON(200, gin.H{
				"result":  "0",
				"message": "添加成功",
				"data":    payload,
			})
			//数据库操作，添加
			db.Create(&payload)
		} else {
			context.JSON(400, gin.H{
				"result":  "1",
				"message": "添加失败",
				"data": gin.H{
					"err": err,
				},
			})
		}
	})
	//DELETE
	ginAPI.DELETE("/user/:id", func(context *gin.Context) {
		var payload []List
		//获取地址栏传进来的ID
		//如果地址栏里面是:id 就用context.Param
		//如果地址栏里面是id = 'xx' 就用context.Query
		id := context.Param("id")

		//判断数据库中ID是否存在
		//where根据什么来查找数据  find是根据什么数据类型来返回数据/什么东西来装填数据
		db.Where("id =?", id).Find(&payload)
		//如果找到了该数据
		if len(payload) > 0 {
			context.JSON(200, gin.H{
				"result": 0,
				"data":   payload,
				"msg":    "删除成功",
			})
			//	数据库删除
			db.Where("id= ?", id).Delete(&payload)

		} else {
			context.JSON(200, gin.H{
				"result": 1,
				"msg":    "删除失败",
			})
		}

	})
	//PUT接口
	ginAPI.PUT("/user/:id", func(context *gin.Context) {
		var data List

		//地址上面的传进来的参数
		id := context.Param("id")
		//查找符合ID筛选条件的数据
		db.Select("id").Where("id = ?", id).Find(&data)
		//判断id是否存在
		if data.ID != 0 {

			err := context.ShouldBindJSON(&data)
			//如果传入的数据符合接口的标准 / 没报错
			if err == nil {

				//修改数据库
				result := db.Where("id = ?", id).Updates(&data)
				rowsAffected := result.RowsAffected
				context.JSON(200, gin.H{
					"result":       "0",
					"msg":          "修改成功",
					"rowsAffected": rowsAffected,
				})
			} else {
				context.JSON(200, gin.H{
					"result": "1",
					"msg":    "传入的数据类型错误",
				})
			}

		} else {
			context.JSON(200, gin.H{
				"result": '1',
				"msg":    "用户ID未找到",
			})
		}
	})

	//注意这个端口数字前面有个 : 符号
	port := ":3001"
	//运行端口
	ginAPI.Run(port)
}
