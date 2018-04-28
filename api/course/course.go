package course

import (
	"github.com/kataras/iris"
	"github.com/mohuishou/scuplus-go/api"
	"github.com/mohuishou/scuplus-go/middleware"
	"github.com/mohuishou/scuplus-go/model"
)

// GetParams Get 参数
type GetParams struct {
	CallName string `form:"call_name"` // 点名方式
	ExamType string `form:"exam_type"` // 考核方式
	Task     string `form:"task"`      // 有无作业
	Day      string `form:"day"`       // 周几上课
	Campus   string `form:"campus"`    // 校区
	Order    string `form:"order"`     // 排序
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// GetCourses 获取课程列表
func GetCourses(ctx iris.Context) {
	params := GetParams{}
	ctx.ReadForm(&params)
	var courseCounts []model.CourseCount
	scope := model.DB().Offset((params.Page - 1) * params.PageSize).Limit(params.PageSize)
	if params.CallName != "" {
		scope = scope.Where("call_name = ?", params.CallName)
	}
	if params.Day != "" {
		scope = scope.Where("day = ?", params.Day)
	}
	if params.ExamType != "" {
		scope = scope.Where("exam_type = ?", params.ExamType)
	}
	if params.Task != "" {
		scope = scope.Where("task = ?", params.Task)
	}
	if params.Campus != "" {
		scope = scope.Where("campus = ?", params.Campus)
	}
	if params.Order != "" {
		scope = scope.Order(params.Order)
	} else {
		scope = scope.Order("avg_grade desc")
	}
	if err := scope.Find(&courseCounts).Error; err != nil {
		api.Error(ctx, 70001, "获取错误", nil)
		return
	}
	api.Success(ctx, "获取成功！", courseCounts)
}

// SearchParams Get 参数
type SearchParams struct {
	Name     string `form:"name"` // 搜索的课程名
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// Search 课程搜索
func Search(ctx iris.Context) {
	params := SearchParams{}
	ctx.ReadForm(&params)
	if params.Name == "" {
		api.Error(ctx, 70400, "参数错误", nil)
		return
	}
	var courseCounts []model.CourseCount
	scope := model.DB().Offset((params.Page - 1) * params.PageSize).Limit(params.PageSize).Order("avg_grade desc")
	if err := scope.Where("name like ?", "%"+params.Name+"%").Find(&courseCounts).Error; err != nil {
		api.Error(ctx, 70001, "获取错误", nil)
		return
	}
	api.Success(ctx, "获取成功！", courseCounts)
}

// Get 获取一门课程的所有信息
// 包括CourseCount\Cousre\CousreEva中的所有信息
func Get(ctx iris.Context) {
	courseID := ctx.URLParam("course_id")
	lessonID := ctx.URLParam("lesson_id")
	if courseID == "" || lessonID == "" {
		api.Error(ctx, 70400, "参数错误", nil)
		return
	}

	// 获取课程信息
	var (
		courseCount     model.CourseCount
		courses         []model.Course
		courseEvaluates []model.CourseEvaluate
	)
	scope := model.DB().Where("course_id = ? and lesson_id = ?", courseID, lessonID)
	scope.Find(&courseCount)
	scope.Find(&courseEvaluates)
	scope.Find(&courses)

	// 获取用户是否有该门课程
	uid := middleware.GetUserID(ctx)
	has := !model.DB().Model(&model.Schedule{}).Where("user_id = ? and course_id = ? and lesson_id = ?", uid, courseID, lessonID).RecordNotFound()
	api.Success(ctx, "获取成功！", map[string]interface{}{
		"course_count":    courseCount,
		"courses":         courses,
		"course_evalutes": courseEvaluates,
		"has":             has, // true: 拥有该门课程, false: 不拥有
	})
}

// CommentParam 课程评价参数
type CommentParam struct {
	ID       uint   `form:"id"`
	CallName int    `form:"call_name"` // 点名方式
	ExamType int    `form:"exam_type"` // 考核方式
	Task     int    `form:"task"`      // 有无作业
	Star     int    `form:"star"`
	CourseID string `form:"course_id"`
	LessonID string `form:"lesson_id"`
	Comment  string `form:"comment"`
}

// Comment 课程评价，目前只能评价正在上的课程
func Comment(ctx iris.Context) {
	params := CommentParam{}
	ctx.ReadForm(&params)

	// 获取用户是否有该门课程
	uid := middleware.GetUserID(ctx)
	hasRecord := model.DB().Where("user_id = ? and course_id = ? and lesson_id = ?", uid, params.CourseID, params.LessonID)
	if hasRecord.Model(&model.Schedule{}).RecordNotFound() {
		api.Error(ctx, 70401, "您的课程表没有该课程！", nil)
		return
	}

	// 检查用户是否已经评价过
	if !hasRecord.Model(&model.CourseEvaluate{}).RecordNotFound() {
		api.Error(ctx, 70401, "该课程您已评价！", nil)
		return
	}

	// 添加课程评价
	courseEvaluate := model.CourseEvaluate{
		CallName: params.CallName,
		ExamType: params.ExamType,
		Task:     params.Task,
		CourseID: params.CourseID,
		LessonID: params.LessonID,
		Comment:  params.Comment,
		UserID:   uid,
		Score:    1,
	}
	if err := model.DB().Create(&courseEvaluate).Error; err != nil {
		api.Error(ctx, 70002, "评教失败！", err)
		return
	}
	api.Success(ctx, "评教成功！", nil)
}

// UpdateComment 更新评价
func UpdateComment(ctx iris.Context) {
	params := CommentParam{}
	ctx.ReadForm(&params)
	if params.ID == 0 {
		api.Error(ctx, 70400, "参数错误！", nil)
		return
	}

	courseEvaluate := model.CourseEvaluate{
		CallName: params.CallName,
		ExamType: params.ExamType,
		Task:     params.Task,
		CourseID: params.CourseID,
		LessonID: params.LessonID,
		Comment:  params.Comment,
	}
	if err := model.DB().Model(&model.CourseEvaluate{
		Model: model.Model{ID: params.ID},
	}).Updates(courseEvaluate).Error; err != nil {
		api.Error(ctx, 70002, "更新失败！", err)
		return
	}
	api.Success(ctx, "更新成功！", nil)
}

// GetComment 获取已经评价的课程
func GetComment(ctx iris.Context) {
	id, err := ctx.URLParamInt("id")
	if id == 0 || err != nil {
		api.Error(ctx, 70400, "参数错误！", nil)
		return
	}
	courseEvaluate := model.CourseEvaluate{}
	if err := model.DB().Find(&courseEvaluate, id); err != nil {
		api.Error(ctx, 70003, "获取失败！", err)
		return
	}

	if courseEvaluate.UserID != middleware.GetUserID(ctx) {
		api.Error(ctx, 70401, "您没有这个权限！", err)
		return
	}

	api.Success(ctx, "获取成功！", courseEvaluate)
}
