/*
* @Author: yangdihang
* @Date: 2020/6/2
 */

package controllers

import (
	"slime.io/slime/framework/model"
	modmodel "slime.io/slime/modules/lazyload/model"
)

var log = modmodel.ModuleLog.WithField(model.LogFieldKeyPkg, "controllers")

type Diff struct {
	Deleted []string
	Added   []string
}
