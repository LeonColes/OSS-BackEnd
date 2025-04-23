// Package docs 提供Swagger文档
package docs

import "github.com/swaggo/swag"

func init() {
	swag.Register(swag.Name, &swag.Spec{
		InfoInstanceName: "swagger",
		SwaggerTemplate:  docTemplate,
	})
}

const docTemplate = `{
  "swagger": "2.0",
  "info": {
    "description": "对象存储系统后端API",
    "title": "OSS-Backend API",
    "contact": {},
    "version": "1.0"
  },
  "host": "localhost:8080",
  "basePath": "/api/oss",
  "paths": {
    "/auth/login": {
      "post": {
        "description": "用户登录接口",
        "consumes": [
          "application/json"
        ],
        "produces": [
          "application/json"
        ],
        "tags": [
          "auth"
        ],
        "summary": "用户登录",
        "parameters": [
          {
            "description": "登录信息",
            "name": "credentials",
            "in": "body",
            "required": true,
            "schema": {
              "type": "object",
              "required": [
                "username",
                "password"
              ],
              "properties": {
                "password": {
                  "type": "string"
                },
                "username": {
                  "type": "string"
                }
              }
            }
          }
        ],
        "responses": {
          "200": {
            "description": "登录成功",
            "schema": {
              "type": "object"
            }
          },
          "400": {
            "description": "参数错误",
            "schema": {
              "type": "object"
            }
          },
          "401": {
            "description": "认证失败",
            "schema": {
              "type": "object"
            }
          }
        }
      }
    },
    "/auth/refresh": {
      "post": {
        "description": "刷新访问令牌",
        "consumes": [
          "application/json"
        ],
        "produces": [
          "application/json"
        ],
        "tags": [
          "auth"
        ],
        "summary": "刷新令牌",
        "parameters": [
          {
            "description": "刷新令牌",
            "name": "refresh_token",
            "in": "body",
            "required": true,
            "schema": {
              "type": "object",
              "required": [
                "refresh_token"
              ],
              "properties": {
                "refresh_token": {
                  "type": "string"
                }
              }
            }
          }
        ],
        "responses": {
          "200": {
            "description": "刷新成功",
            "schema": {
              "type": "object"
            }
          },
          "400": {
            "description": "参数错误",
            "schema": {
              "type": "object"
            }
          },
          "401": {
            "description": "令牌无效",
            "schema": {
              "type": "object"
            }
          }
        }
      }
    },
    "/auth/register": {
      "post": {
        "description": "用户注册接口",
        "consumes": [
          "application/json"
        ],
        "produces": [
          "application/json"
        ],
        "tags": [
          "auth"
        ],
        "summary": "用户注册",
        "parameters": [
          {
            "description": "注册信息",
            "name": "user",
            "in": "body",
            "required": true,
            "schema": {
              "type": "object",
              "required": [
                "username",
                "password",
                "email"
              ],
              "properties": {
                "email": {
                  "type": "string"
                },
                "password": {
                  "type": "string"
                },
                "username": {
                  "type": "string"
                }
              }
            }
          }
        ],
        "responses": {
          "201": {
            "description": "注册成功",
            "schema": {
              "type": "object"
            }
          },
          "400": {
            "description": "参数错误",
            "schema": {
              "type": "object"
            }
          },
          "409": {
            "description": "用户已存在",
            "schema": {
              "type": "object"
            }
          }
        }
      }
    },
    "/file/{id}": {
      "get": {
        "security": [
          {
            "BearerAuth": []
          }
        ],
        "description": "获取指定文件的详细信息",
        "produces": [
          "application/json"
        ],
        "tags": [
          "file"
        ],
        "summary": "获取文件信息",
        "parameters": [
          {
            "type": "integer",
            "description": "文件ID",
            "name": "id",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "200": {
            "description": "获取成功",
            "schema": {
              "type": "object"
            }
          },
          "400": {
            "description": "请求参数错误",
            "schema": {
              "type": "object"
            }
          },
          "401": {
            "description": "未授权",
            "schema": {
              "type": "object"
            }
          },
          "404": {
            "description": "文件不存在",
            "schema": {
              "type": "object"
            }
          },
          "500": {
            "description": "服务器内部错误",
            "schema": {
              "type": "object"
            }
          }
        }
      },
      "delete": {
        "security": [
          {
            "BearerAuth": []
          }
        ],
        "description": "删除指定的文件或文件夹",
        "produces": [
          "application/json"
        ],
        "tags": [
          "file"
        ],
        "summary": "删除文件/文件夹",
        "parameters": [
          {
            "type": "integer",
            "description": "文件ID",
            "name": "id",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "200": {
            "description": "删除成功",
            "schema": {
              "type": "object"
            }
          },
          "400": {
            "description": "请求参数错误",
            "schema": {
              "type": "object"
            }
          },
          "401": {
            "description": "未授权",
            "schema": {
              "type": "object"
            }
          },
          "403": {
            "description": "无权操作",
            "schema": {
              "type": "object"
            }
          },
          "404": {
            "description": "文件不存在",
            "schema": {
              "type": "object"
            }
          },
          "409": {
            "description": "文件夹不为空",
            "schema": {
              "type": "object"
            }
          },
          "500": {
            "description": "服务器内部错误",
            "schema": {
              "type": "object"
            }
          }
        }
      }
    },
    "/file/{id}/download": {
      "get": {
        "security": [
          {
            "BearerAuth": []
          }
        ],
        "description": "下载指定文件",
        "produces": [
          "application/octet-stream"
        ],
        "tags": [
          "file"
        ],
        "summary": "下载文件",
        "parameters": [
          {
            "type": "integer",
            "description": "文件ID",
            "name": "id",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "200": {
            "description": "文件内容",
            "schema": {
              "type": "file"
            }
          },
          "400": {
            "description": "请求参数错误",
            "schema": {
              "type": "object"
            }
          },
          "401": {
            "description": "未授权",
            "schema": {
              "type": "object"
            }
          },
          "403": {
            "description": "无权操作",
            "schema": {
              "type": "object"
            }
          },
          "404": {
            "description": "文件不存在",
            "schema": {
              "type": "object"
            }
          },
          "500": {
            "description": "服务器内部错误",
            "schema": {
              "type": "object"
            }
          }
        }
      }
    },
    "/file/{id}/move": {
      "put": {
        "security": [
          {
            "BearerAuth": []
          }
        ],
        "description": "移动文件或文件夹到指定目录",
        "consumes": [
          "application/json"
        ],
        "produces": [
          "application/json"
        ],
        "tags": [
          "file"
        ],
        "summary": "移动文件/文件夹",
        "parameters": [
          {
            "type": "integer",
            "description": "文件ID",
            "name": "id",
            "in": "path",
            "required": true
          },
          {
            "description": "移动信息",
            "name": "request",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/file.MoveFileRequest"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "移动成功",
            "schema": {
              "type": "object"
            }
          },
          "400": {
            "description": "请求参数错误",
            "schema": {
              "type": "object"
            }
          },
          "401": {
            "description": "未授权",
            "schema": {
              "type": "object"
            }
          },
          "403": {
            "description": "无权操作",
            "schema": {
              "type": "object"
            }
          },
          "404": {
            "description": "文件不存在",
            "schema": {
              "type": "object"
            }
          },
          "409": {
            "description": "目标路径已存在同名文件",
            "schema": {
              "type": "object"
            }
          },
          "500": {
            "description": "服务器内部错误",
            "schema": {
              "type": "object"
            }
          }
        }
      }
    },
    "/file/{id}/rename": {
      "put": {
        "security": [
          {
            "BearerAuth": []
          }
        ],
        "description": "重命名指定的文件或文件夹",
        "consumes": [
          "application/json"
        ],
        "produces": [
          "application/json"
        ],
        "tags": [
          "file"
        ],
        "summary": "重命名文件/文件夹",
        "parameters": [
          {
            "type": "integer",
            "description": "文件ID",
            "name": "id",
            "in": "path",
            "required": true
          },
          {
            "description": "重命名信息",
            "name": "request",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/file.RenameFileRequest"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "重命名成功",
            "schema": {
              "type": "object"
            }
          },
          "400": {
            "description": "请求参数错误",
            "schema": {
              "type": "object"
            }
          },
          "401": {
            "description": "未授权",
            "schema": {
              "type": "object"
            }
          },
          "403": {
            "description": "无权操作",
            "schema": {
              "type": "object"
            }
          },
          "404": {
            "description": "文件不存在",
            "schema": {
              "type": "object"
            }
          },
          "409": {
            "description": "文件名已存在",
            "schema": {
              "type": "object"
            }
          },
          "500": {
            "description": "服务器内部错误",
            "schema": {
              "type": "object"
            }
          }
        }
      }
    },
    "/file/folder": {
      "post": {
        "security": [
          {
            "BearerAuth": []
          }
        ],
        "description": "在指定项目和路径下创建文件夹",
        "consumes": [
          "application/json"
        ],
        "produces": [
          "application/json"
        ],
        "tags": [
          "file"
        ],
        "summary": "创建文件夹",
        "parameters": [
          {
            "description": "文件夹信息",
            "name": "request",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/file.CreateFolderRequest"
            }
          }
        ],
        "responses": {
          "201": {
            "description": "创建成功",
            "schema": {
              "type": "object"
            }
          },
          "400": {
            "description": "请求参数错误",
            "schema": {
              "type": "object"
            }
          },
          "401": {
            "description": "未授权",
            "schema": {
              "type": "object"
            }
          },
          "403": {
            "description": "无权操作",
            "schema": {
              "type": "object"
            }
          },
          "409": {
            "description": "文件夹已存在",
            "schema": {
              "type": "object"
            }
          },
          "500": {
            "description": "服务器内部错误",
            "schema": {
              "type": "object"
            }
          }
        }
      }
    },
    "/file/list": {
      "get": {
        "security": [
          {
            "BearerAuth": []
          }
        ],
        "description": "获取指定项目和路径下的文件列表",
        "produces": [
          "application/json"
        ],
        "tags": [
          "file"
        ],
        "summary": "获取文件列表",
        "parameters": [
          {
            "type": "integer",
            "description": "项目ID",
            "name": "project_id",
            "in": "query",
            "required": true
          },
          {
            "type": "string",
            "description": "文件夹路径",
            "name": "folder_path",
            "in": "query"
          },
          {
            "type": "integer",
            "default": 1,
            "description": "页码",
            "name": "page",
            "in": "query"
          },
          {
            "type": "integer",
            "default": 20,
            "description": "每页数量",
            "name": "page_size",
            "in": "query"
          },
          {
            "type": "boolean",
            "default": false,
            "description": "是否显示已删除文件",
            "name": "show_deleted",
            "in": "query"
          }
        ],
        "responses": {
          "200": {
            "description": "获取成功",
            "schema": {
              "type": "object"
            }
          },
          "400": {
            "description": "请求参数错误",
            "schema": {
              "type": "object"
            }
          },
          "401": {
            "description": "未授权",
            "schema": {
              "type": "object"
            }
          },
          "403": {
            "description": "无权操作",
            "schema": {
              "type": "object"
            }
          },
          "500": {
            "description": "服务器内部错误",
            "schema": {
              "type": "object"
            }
          }
        }
      }
    },
    "/file/share": {
      "post": {
        "security": [
          {
            "BearerAuth": []
          }
        ],
        "description": "创建文件共享链接",
        "consumes": [
          "application/json"
        ],
        "produces": [
          "application/json"
        ],
        "tags": [
          "file"
        ],
        "summary": "创建文件共享",
        "parameters": [
          {
            "description": "共享信息",
            "name": "request",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/file.CreateShareRequest"
            }
          }
        ],
        "responses": {
          "201": {
            "description": "创建成功",
            "schema": {
              "type": "object"
            }
          },
          "400": {
            "description": "请求参数错误",
            "schema": {
              "type": "object"
            }
          },
          "401": {
            "description": "未授权",
            "schema": {
              "type": "object"
            }
          },
          "403": {
            "description": "无权操作",
            "schema": {
              "type": "object"
            }
          },
          "404": {
            "description": "文件不存在",
            "schema": {
              "type": "object"
            }
          },
          "500": {
            "description": "服务器内部错误",
            "schema": {
              "type": "object"
            }
          }
        }
      }
    },
    "/file/upload": {
      "post": {
        "security": [
          {
            "BearerAuth": []
          }
        ],
        "description": "上传文件到指定项目和路径",
        "consumes": [
          "multipart/form-data"
        ],
        "produces": [
          "application/json"
        ],
        "tags": [
          "file"
        ],
        "summary": "上传文件",
        "parameters": [
          {
            "type": "integer",
            "description": "项目ID",
            "name": "project_id",
            "in": "formData",
            "required": true
          },
          {
            "type": "string",
            "description": "文件夹路径",
            "name": "folder_path",
            "in": "formData",
            "required": true
          },
          {
            "type": "file",
            "description": "要上传的文件",
            "name": "file",
            "in": "formData",
            "required": true
          }
        ],
        "responses": {
          "201": {
            "description": "上传成功",
            "schema": {
              "type": "object"
            }
          },
          "400": {
            "description": "请求参数错误",
            "schema": {
              "type": "object"
            }
          },
          "401": {
            "description": "未授权",
            "schema": {
              "type": "object"
            }
          },
          "403": {
            "description": "无权操作",
            "schema": {
              "type": "object"
            }
          },
          "500": {
            "description": "服务器内部错误",
            "schema": {
              "type": "object"
            }
          }
        }
      }
    },
    "/share/{code}": {
      "get": {
        "description": "获取指定共享码的文件共享信息",
        "produces": [
          "application/json"
        ],
        "tags": [
          "file"
        ],
        "summary": "获取共享信息",
        "parameters": [
          {
            "type": "string",
            "description": "共享码",
            "name": "code",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "200": {
            "description": "获取成功",
            "schema": {
              "type": "object"
            }
          },
          "400": {
            "description": "请求参数错误",
            "schema": {
              "type": "object"
            }
          },
          "404": {
            "description": "共享不存在或已过期",
            "schema": {
              "type": "object"
            }
          },
          "500": {
            "description": "服务器内部错误",
            "schema": {
              "type": "object"
            }
          }
        }
      }
    }
  },
  "definitions": {
    "file.CreateFolderRequest": {
      "type": "object",
      "required": [
        "folder_name",
        "project_id"
      ],
      "properties": {
        "folder_name": {
          "type": "string"
        },
        "folder_path": {
          "type": "string"
        },
        "project_id": {
          "type": "integer"
        }
      }
    },
    "file.CreateShareRequest": {
      "type": "object",
      "required": [
        "file_id"
      ],
      "properties": {
        "download_limit": {
          "type": "integer"
        },
        "expire_hours": {
          "type": "integer"
        },
        "file_id": {
          "type": "integer"
        },
        "password": {
          "type": "string"
        }
      }
    },
    "file.MoveFileRequest": {
      "type": "object",
      "required": [
        "target_path"
      ],
      "properties": {
        "target_path": {
          "type": "string"
        }
      }
    },
    "file.RenameFileRequest": {
      "type": "object",
      "required": [
        "new_name"
      ],
      "properties": {
        "new_name": {
          "type": "string"
        }
      }
    }
  },
  "securityDefinitions": {
    "BearerAuth": {
      "type": "apiKey",
      "name": "Authorization",
      "in": "header"
    }
  }
}` 