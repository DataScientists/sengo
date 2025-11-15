package controller

// Controller struct holds the controller of the entire app
type Controller struct {
	User         interface{ User }
	Todo         interface{ Todo }
	Profile      interface{ Profile }
	Auth         interface{ Auth }
	ProfileEntry interface{ ProfileEntry }
	APIQuota     interface{ APIQuota }
	CronJob      interface{ CronJob }
	JobExecution interface{ JobExecution }
	Dashboard    interface{ Dashboard }
}
