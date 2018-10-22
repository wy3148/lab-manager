package controllers

import (
	"github.com/astaxie/beego"
	"github.com/wy3148/lab-manager/labm/models"
	"github.com/wy3148/lab-manager/labm/util"
)

type DockerController struct {
	beego.Controller
}

func (d *DockerController) Post() {
	uid := d.GetString("user")
	if len(uid) == 0 {
		util.Log.Error("Empty user id")
	}

	j, err := models.NewJupyterDocker(uid)
	if err != nil {
		d.Ctx.ResponseWriter.Status = 500
		d.Ctx.ResponseWriter.Write([]byte("internal error"))
		return
	}

	d.Data["json"] = j
	d.ServeJSON()
}

func (d *DockerController) Get() {
	uid := d.GetString("user")
	ex := func() {
		d.Ctx.ResponseWriter.Status = 500
		d.Ctx.ResponseWriter.Write([]byte("internal error"))
	}

	if len(uid) == 0 {
		util.Log.Error("Empty user id")
		ex()
		return
	}

	j, err := models.NewJupyterDocker(uid)
	if err != nil {
		ex()
		return
	}

	d.Data["json"] = j
	d.ServeJSON()
}
