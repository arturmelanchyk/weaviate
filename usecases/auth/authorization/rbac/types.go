//                           _       _
// __      _____  __ ___   ___  __ _| |_ ___
// \ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
//  \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
//   \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
//
//  Copyright © 2016 - 2024 Weaviate B.V. All rights reserved.
//
//  CONTACT: hello@weaviate.io
//

package rbac

import (
	"fmt"
	"strings"

	"github.com/go-openapi/strfmt"
	"github.com/weaviate/weaviate/entities/models"
	"github.com/weaviate/weaviate/usecases/auth/authorization"
)

const (
	rolesD            = "roles"
	cluster           = "cluster"
	collections       = "collections"
	tenants           = "tenants"
	objectsCollection = "objects_collection"
	objectsTenants    = "objects_tenants"

	// rolePrefix = "r_"
	// userPrefix = "u_"
)

var builtInRoles = map[string]string{
	"viewer": authorization.READ,
	"editor": authorization.CRU,
	"admin":  authorization.CRUD,
}

func policy(permission *models.Permission) (string, string, string, error) {
	// TODO verify slice position to avoid panics
	action, domain, found := strings.Cut(*permission.Action, "_")
	if !found {
		return "", "", "", fmt.Errorf("invalid action")
	}
	verb := strings.ToUpper(action[:1])
	if verb == "M" {
		verb = authorization.CRUD
	}
	var resource string
	switch domain {
	case rolesD:
		resource = authorization.Roles()[0]
	case cluster:
		resource = authorization.Cluster()
	case collections:
		if permission.Collection == nil {
			resource = authorization.Collections("*")[0]
		} else {
			resource = authorization.Collections(*permission.Collection)[0]
		}
	case tenants:
		if permission.Collection == nil || permission.Tenant == nil {
			resource = authorization.Shards("*", "*")[0]
		} else {
			resource = authorization.Shards(*permission.Collection, *permission.Tenant)[0]
		}
	case objectsCollection, objectsTenants:
		if permission.Collection == nil || permission.Tenant == nil || permission.Object == nil {
			resource = authorization.Objects("*", "*", "*")
		} else {
			resource = authorization.Objects(*permission.Collection, *permission.Tenant, strfmt.UUID(*permission.Object))
		}
	}
	return resource, verb, domain, nil
}

func permission(policy []string) *models.Permission {
	domain := policy[3]
	action := fmt.Sprintf("%s_%s", authorization.Actions[policy[2]], domain)

	action = strings.ReplaceAll(action, "_*", "")
	permission := &models.Permission{
		Action: &action,
	}

	splits := strings.Split(policy[1], "/")
	all := "*"

	switch domain {
	case collections:
		permission.Collection = &splits[1]
	case tenants:
		permission.Tenant = &splits[3]
		permission.Collection = &splits[1]
	case objectsTenants, objectsCollection:
		permission.Collection = &splits[1]
		permission.Tenant = &splits[3]
		permission.Object = &splits[5]
	case rolesD:
		permission.Role = &splits[1]
	// case cluster:

	case "*":
		permission.Collection = &all
		permission.Tenant = &all
		permission.Object = &all
		permission.Role = &all
	}

	return permission
}
