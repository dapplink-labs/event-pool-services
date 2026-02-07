-- ============================================
-- 事件池统数据库初始化脚本
-- Event Services System Database Schema
-- ============================================

-- 创建自定义类型：UINT256（无符号 256 位整数）
DO
$$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'uint256') THEN
        CREATE DOMAIN UINT256 AS NUMERIC CHECK (VALUE >= 0 AND VALUE < POWER(CAST(2 AS NUMERIC), CAST(256 AS NUMERIC)) AND SCALE(VALUE) = 0);
    ELSE
        ALTER DOMAIN UINT256 DROP CONSTRAINT uint256_check;
        ALTER DOMAIN UINT256 ADD CHECK (VALUE >= 0 AND VALUE < POWER(CAST(2 AS NUMERIC), CAST(256 AS NUMERIC)) AND SCALE(VALUE) = 0);
    END IF;
END
$$;

-- 启用 UUID 扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp" CASCADE;

-- ============================================
-- 系统表区域 (Vben Admin Compatible + Merged Business Logic)
-- ============================================
-- 系统菜单表 (sys_menus) - Replaces 'auth'
CREATE TABLE IF NOT EXISTS sys_menus (
    guid                  TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    parent_id             VARCHAR(255) DEFAULT '',    -- 父菜单ID
    tree_path             VARCHAR(1024) DEFAULT '',   -- 父节点ID路径
    name                  VARCHAR(255) DEFAULT '',    -- 菜单名称/标题 （对应前端中文字段 Meta.Title）
    type                  VARCHAR(50) DEFAULT '',     -- 菜单类型：0=目录，1=菜单，2=按钮
    route_name            VARCHAR(255) DEFAULT '',    -- 路由名称 (对应前端英文字段 Name)
    path                  VARCHAR(255) DEFAULT '',    -- 路由路径
    component             VARCHAR(255) DEFAULT '',    -- 组件路径
    perm                  VARCHAR(255) DEFAULT '',    -- 权限标识 (原 auth.auth_url/auth_name)
    status                SMALLINT DEFAULT 1,         -- 状态：1=启用，2=禁用
    sort                  BIGINT DEFAULT 0,           -- 排序
    icon                  VARCHAR(255) DEFAULT '',    -- 图标
    redirect              VARCHAR(255) DEFAULT '',    -- 跳转路径
    -- Vben Meta 字段
    affix_tab             SMALLINT DEFAULT 2,         -- 固定标签页 (1是 2否)
    hide_children_in_menu SMALLINT DEFAULT 2,         -- 子级不展现 (1是 2否)
    hide_in_breadcrumb    SMALLINT DEFAULT 2,         -- 面包屑中不展现 (1是 2否)
    hide_in_menu          SMALLINT DEFAULT 2,         -- 菜单中不展现 (原 auth.is_show) (1是 2否)
    hide_in_tab           SMALLINT DEFAULT 2,         -- 标签页中不展现 (1是 2否)
    keep_alive            SMALLINT DEFAULT 1,         -- 是否缓存 (1是 2否)
    created_at            TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_sys_menus_parent_id ON sys_menus(parent_id);
CREATE INDEX IF NOT EXISTS idx_sys_menus_status ON sys_menus(status);

-- 系统API接口表 (sys_apis)
CREATE TABLE IF NOT EXISTS sys_apis (
    guid                  TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    parent_id             VARCHAR(255) DEFAULT '',    -- 父级ID
    path                  VARCHAR(255) DEFAULT '',    -- API路径
    description           VARCHAR(255) DEFAULT '',    -- 描述
    method                VARCHAR(20) DEFAULT '',     -- 方法 (GET, POST...)
    created_at            TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_sys_apis_parent_id ON sys_apis(parent_id);
CREATE INDEX IF NOT EXISTS idx_sys_apis_path_method ON sys_apis(path, method);

-- 系统角色表 (sys_roles) - Replaces 'role'
CREATE TABLE IF NOT EXISTS sys_roles (
    guid                  TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    name                  VARCHAR(255) DEFAULT '',    -- 角色名称 (原 role_name)
    code                  VARCHAR(255) DEFAULT '',    -- 角色编码
    sort                  BIGINT DEFAULT 0,           -- 排序
    status                SMALLINT DEFAULT 1,         -- 状态
    remark                VARCHAR(255) DEFAULT '',    -- 备注 (原 detail)
    created_at            TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_sys_roles_code ON sys_roles(code) WHERE deleted_at IS NULL;

-- 系统用户表 (sys_users) - Replaces 'admin'
CREATE TABLE IF NOT EXISTS sys_users (
    guid                  TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    username              VARCHAR(255) NOT NULL,      -- 用户名 (原 login_name)
    nickname              VARCHAR(255) DEFAULT '',    -- 昵称 (原 real_name)
    password              VARCHAR(255) DEFAULT '',    -- 密码
    salt                  VARCHAR(255) DEFAULT '',    -- 盐
    mobile                VARCHAR(20) DEFAULT '',     -- 手机 (原 phone)
    email                 VARCHAR(255) DEFAULT '',    -- 邮箱
    avatar                VARCHAR(500) DEFAULT '',    -- 头像
    gender                SMALLINT DEFAULT 0,         -- 性别 1男 2女 0未知
    status                SMALLINT DEFAULT 1,         -- 状态 1正常 2禁用
    dept_id               VARCHAR(255) DEFAULT '',    -- 部门ID
    role_id               VARCHAR(255) DEFAULT '',    -- 角色ID (主角色, 原 role_ids)
    remark                VARCHAR(500) DEFAULT '',    -- 备注
    create_by             VARCHAR(255) DEFAULT '',    -- 创建者ID
    update_by             VARCHAR(255) DEFAULT '',    -- 更新者ID
    last_login_time       BIGINT DEFAULT 0,           -- 最后登录时间 (原 last_login)
    last_login_ip         VARCHAR(255) DEFAULT '',    -- 最后登录IP (原 last_ip)
    created_at            TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_sys_users_username ON sys_users(username) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_sys_users_role_id ON sys_users(role_id);

-- 角色与菜单关联表 (sys_role_auths) - Replaces 'role_auth'
CREATE TABLE IF NOT EXISTS sys_role_auths (
    role_id               VARCHAR(255) NOT NULL,
    auth_id               VARCHAR(255) NOT NULL, -- 对应 sys_menus.guid
    PRIMARY KEY (role_id, auth_id)
);
CREATE INDEX IF NOT EXISTS idx_sys_role_auths_auth_id ON sys_role_auths(auth_id);

-- 角色与API关联表 (sys_role_apis)
CREATE TABLE IF NOT EXISTS sys_role_apis (
    role_id               VARCHAR(255) NOT NULL,
    api_id                VARCHAR(255) NOT NULL, -- 对应 sys_apis.guid
    PRIMARY KEY (role_id, api_id)
);

-- 数据字典表 (sys_dicts)
CREATE TABLE IF NOT EXISTS sys_dicts (
    guid                  TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    dict_name             VARCHAR(255) DEFAULT '',    -- 字典名称
    dict_type             VARCHAR(255) DEFAULT '',    -- 字典类型
    item_key              VARCHAR(255) DEFAULT '',    -- 键
    item_value            VARCHAR(255) DEFAULT '',    -- 值
    sort                  BIGINT DEFAULT 0,           -- 排序
    status                SMALLINT DEFAULT 1,         -- 状态
    remark                VARCHAR(255) DEFAULT '',    -- 备注
    created_at            TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_sys_dicts_type ON sys_dicts(dict_type);

-- 操作日志表 (sys_records) - Replaces 'sys_log'
-- Merges fields from 'sys_log' (business logs) and Vben 'sys_records'
CREATE TABLE IF NOT EXISTS sys_records (
    guid                  TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    -- Vben Standard Fields
    username              VARCHAR(255) DEFAULT '',    -- 操作人用户名
    user_id               VARCHAR(255) DEFAULT '',    -- 操作人ID (原 admin_id)
    description           VARCHAR(500) DEFAULT '',    -- 描述 (原 sys_log.desc)
    method                VARCHAR(20) DEFAULT '',     -- 请求方法
    path                  VARCHAR(500) DEFAULT '',    -- 请求路径 (原 sys_log.action)
    status_code           INTEGER DEFAULT 0,          -- HTTP状态码
    elapsed               VARCHAR(50) DEFAULT '',     -- 耗时
    msg                   TEXT DEFAULT '',            -- 消息
    request               TEXT DEFAULT '',            -- 请求体
    response              TEXT DEFAULT '',            -- 响应体
    platform              VARCHAR(50) DEFAULT '',     -- 平台
    ip                    VARCHAR(50) DEFAULT '',     -- IP地址 (原 sys_log.ip)
    address               VARCHAR(500) DEFAULT '',    -- 物理地址

-- Business/Merged Fields from sys_log
    cate                  SMALLINT DEFAULT 0,         -- 日志类型：0=其他, 1=登录, 2=财务操作
    status                SMALLINT DEFAULT -1,        -- 业务状态：0=成功, 1=失败, -1=未知 (Differ from HTTP status_code)
    asset                 VARCHAR(255) DEFAULT '',    -- 相关资产/币种
    before_val            VARCHAR(255) DEFAULT '',    -- 修改前的值 (field renamed from 'before')
    after_val             VARCHAR(255) DEFAULT '',    -- 修改后的值 (field renamed from 'after')
    target_user_id        VARCHAR(255) DEFAULT '',    -- 关联目标用户ID (原 sys_log.user_id)
    order_number          VARCHAR(64) DEFAULT '',     -- 关联订单号
    op                    SMALLINT DEFAULT -1,        -- 操作类型：0=添加, 1=编辑, -1=未知
    created_at            TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_sys_records_created_at ON sys_records(created_at);
CREATE INDEX IF NOT EXISTS idx_sys_records_user_id ON sys_records(user_id);
CREATE INDEX IF NOT EXISTS idx_sys_records_target_user_id ON sys_records(target_user_id);
CREATE INDEX IF NOT EXISTS idx_sys_records_cate ON sys_records(cate);
CREATE INDEX IF NOT EXISTS idx_sys_records_order_number ON sys_records(order_number);

-- Casbin 规则表 (casbin_rule)
CREATE TABLE IF NOT EXISTS casbin_rule (
    id                    SERIAL PRIMARY KEY,
    ptype                 VARCHAR(100),
    v0                    VARCHAR(100),
    v1                    VARCHAR(100),
    v2                    VARCHAR(100),
    v3                    VARCHAR(100),
    v4                    VARCHAR(100),
    v5                    VARCHAR(100)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_casbin_rule ON casbin_rule(ptype, v0, v1, v2, v3, v4, v5);

-- ============================================
-- 支持的语言表 (Languages)
-- 定义系统支持的多语言配置
-- ============================================
CREATE TABLE IF NOT EXISTS languages (
    guid             TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),  -- 主键：全局唯一标识符
    language_name    VARCHAR(20) NOT NULL UNIQUE,               -- 语言代码（如：zh、en、ja）- 唯一索引
    language_label   VARCHAR(50),                               -- 语言显示名称（如：中文、English）
    is_default       BOOLEAN NOT NULL DEFAULT FALSE,            -- 是否为默认语言（系统中只能有一个默认语言）
    is_active        BOOLEAN NOT NULL DEFAULT TRUE,             -- 是否启用
    created_at       TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,    -- 创建时间
    updated_at       TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP     -- 更新时间
);
CREATE INDEX IF NOT EXISTS idx_languages_active ON languages (is_active);  -- 按状态查询
CREATE INDEX IF NOT EXISTS idx_languages_default ON languages (is_default);  -- 按默认语言查询
CREATE UNIQUE INDEX IF NOT EXISTS uq_languages_name ON languages (language_name);  -- 语言名称唯一索引
CREATE UNIQUE INDEX IF NOT EXISTS uq_languages_default ON languages (is_default) WHERE is_default = TRUE AND deleted_at IS NULL;  -- 确保只有一个默认语言

-- 分类表 --
-- 关联主键: category_guid --
CREATE TABLE IF NOT EXISTS category (
    guid        TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''), -- 分类主键 GUID（内部关联用）
    code        VARCHAR(64),                                -- 业务编码（稳定标识）；例：SPORTS、LANGUAGES
    sort_order  INTEGER NOT NULL DEFAULT 0,                 -- 排序字段（后台展示/拖拽排序）
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,              -- 是否启用：TRUE=启用；FALSE=下架/禁用
    remark      VARCHAR(200),                               -- 运营备注：仅后台可见
    created_at  TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,     -- 创建时间
    updated_at  TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP      -- 更新时间（建议应用层或触发器维护）
);
CREATE INDEX IF NOT EXISTS idx_category_guid ON category(guid);
CREATE UNIQUE INDEX IF NOT EXISTS uq_category_code ON category(code) WHERE deleted_at IS NULL;  -- 部分唯一索引：只对未删除记录强制唯一
CREATE INDEX IF NOT EXISTS idx_category_active_sort ON category(is_active, sort_order, created_at) WHERE deleted_at IS NULL;

-- 事件分类多语言表 --
CREATE TABLE IF NOT EXISTS category_language (
    guid                   TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    language_guid          VARCHAR(255)  NOT NULL,
    category_guid          VARCHAR(255)  NOT NULL,
    parent_category_guid   VARCHAR(255)  NOT NULL,
    level                  SMALLINT NOT NULL DEFAULT 0, -- 0: 一级分类；1: 二级分类 --
    name                   VARCHAR(50) NOT NULL,
    description            VARCHAR(200) NOT NULL,
    created_at             TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at             TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_category_language_guid ON category_language(guid);
CREATE INDEX IF NOT EXISTS idx_category_language_language_guid ON category_language(language_guid);
CREATE INDEX IF NOT EXISTS idx_category_language_category_guid ON category_language(category_guid);
CREATE INDEX IF NOT EXISTS idx_category_language_parent_category_guid ON category_language(parent_category_guid);
CREATE UNIQUE INDEX IF NOT EXISTS uq_category_language_lang_cat_not_deleted ON category_language(language_guid, category_guid)
    WHERE deleted_at IS NULL; -- 确保同一种 language 下，同一个 category 只能有 1 条

-- 所属生态 --
CREATE TABLE IF NOT EXISTS ecosystem (
    guid            TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    category_guid   VARCHAR(255) NOT NULL,
    event_num       UINT256 NOT NULL CHECK (event_num > 0),     -- 生态的事件个数 --
    code            VARCHAR(64),                                -- 业务编码（稳定标识）；
    sort_order      INTEGER NOT NULL DEFAULT 0,                 -- 排序字段（后台展示/拖拽排序）
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,              -- 是否启用：TRUE=启用；FALSE=下架/禁用
    remark          VARCHAR(200),                               -- 运营备注：仅后台可见
    extra           JSONB,                                      -- 扩展字段（JSON）：临时配置/个性化属性
    created_at      TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,     -- 创建时间
    updated_at      TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP      -- 更新时间（建议应用层或触发器维护）
);
CREATE INDEX IF NOT EXISTS idx_ecosystem_guid ON ecosystem(guid);
CREATE INDEX IF NOT EXISTS idx_ecosystem_category_guid ON ecosystem(category_guid);
CREATE UNIQUE INDEX IF NOT EXISTS uq_ecosystem_code ON ecosystem(code) WHERE deleted_at IS NULL;  -- 部分唯一索引：只对未删除记录强制唯一
CREATE INDEX IF NOT EXISTS idx_ecosystem_active_sort ON ecosystem(is_active, sort_order, created_at) WHERE deleted_at IS NULL;  -- 复合索引：优化列表查询

-- 所属生态多语言表 --
CREATE TABLE IF NOT EXISTS ecosystem_language (
    guid             TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    language_guid    VARCHAR(255)  NOT NULL,
    ecosystem_guid   VARCHAR(255)  NOT NULL,
    name             VARCHAR(50) NOT NULL,
    description      VARCHAR(200) NOT NULL,
    created_at       TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_ecosystem_language_guid ON ecosystem_language(guid);
CREATE INDEX IF NOT EXISTS idx_ecosystem_language_language_guid ON ecosystem_language(language_guid);
CREATE INDEX IF NOT EXISTS idx_ecosystem_language_ecosystem_guid ON ecosystem_language(ecosystem_guid);

-- 时间标签表 --
CREATE TABLE IF NOT EXISTS event_period (
    guid            TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    category_guid   VARCHAR(255) NOT NULL,
    code            VARCHAR(64),                                -- 业务编码（稳定标识）；
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,              -- 是否启用：TRUE=启用；FALSE=下架/禁用
    remark          VARCHAR(200),                               -- 运营备注：仅后台可见
    extra           JSONB,                                      -- 扩展字段（JSON）：临时配置/个性化属性
    created_at      TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,     -- 创建时间
    updated_at      TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP      -- 更新时间（建议应用层或触发器维护）
);
CREATE INDEX IF NOT EXISTS idx_event_period_guid ON event_period(guid);
CREATE INDEX IF NOT EXISTS idx_event_period_category_guid ON event_period(category_guid);
CREATE UNIQUE INDEX IF NOT EXISTS uq_event_period_code ON event_period(code) WHERE deleted_at IS NULL;  -- 部分唯一索引：只对未删除记录强制唯一
CREATE INDEX IF NOT EXISTS idx_event_period_active_sort ON event_period(is_active, sort_order, created_at) WHERE deleted_at IS NULL;  -- 复合索引：优化列表查询

-- 时间标签表多语言表 --
CREATE TABLE IF NOT EXISTS event_period_language (
    guid                TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    event_period_guid   VARCHAR(255)  NOT NULL,
    language_guid       VARCHAR(255)  NOT NULL,
    name                VARCHAR(50) NOT NULL,
    description         VARCHAR(200) NOT NULL,
    created_at          TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_event_period_language_guid ON event_period_language(guid);
CREATE INDEX IF NOT EXISTS idx_event_period_language_event_period_guid ON event_period_language(event_period_guid);
CREATE INDEX IF NOT EXISTS idx_event_period_language_language_guid ON event_period_language(language_guid);

-- 运动类团队 --
CREATE TABLE IF NOT EXISTS team_group (
    guid                TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    logo                VARCHAR(255)  NOT NULL,
    created_at          TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_team_group_guid ON team_group(guid);

CREATE TABLE IF NOT EXISTS team_group_language (
    guid                TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    language_guid       VARCHAR(255)  NOT NULL,
    team_group_guid     VARCHAR(255)  NOT NULL,
    name                VARCHAR(255)  NOT NULL,
    created_at          TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_team_group_language_guid ON team_group_language(guid);
CREATE INDEX IF NOT EXISTS idx_team_group_language_language_guid ON team_group_language(language_guid);
CREATE INDEX IF NOT EXISTS idx_team_group_language_team_group_guid ON team_group_language(team_group_guid);

-- ============================================
-- 事件表 (Event)
-- 存储预测事件的核心信息
-- ============================================
CREATE TABLE IF NOT EXISTS event (
    guid                     TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),  -- 主键：事件唯一标识
    category_guid            VARCHAR(500) NOT NULL,              -- 分类 GUID（外键关联 category 表）
    ecosystem_guid           VARCHAR(500) NOT NULL,              -- 生态 GUID（外键关联 ecosystem 表）
    event_period_guid        VARCHAR(500) NOT NULL,              -- 时间标签 GUID（外键关联 event_period 表）
    main_team_group_guid     VARCHAR(255) NOT NULL DEFAULT '0',  -- 主队 GUID（非运动类为 0）
    cluster_team_group_guid  VARCHAR(255) NOT NULL DEFAULT '0',  -- 客队 GUID（非运动类为 0）
    main_score               UINT256 NOT NULL DEFAULT 0,         -- 主队得分
    cluster_score            UINT256 NOT NULL DEFAULT 0,         -- 客队得分
    logo                     VARCHAR(500) NOT NULL,              -- 事件图标 URL
    event_type               SMALLINT NOT NULL DEFAULT 0,        -- 排序类型：0=热门话题, 1=突发, 2=最新
    experiment_result        TEXT NOT NULL DEFAULT '',           -- 实验结果/事件结果
    info                     JSONB NOT NULL DEFAULT '{}'::jsonb, -- 附加信息（JSON 格式）
    is_online                BOOLEAN NOT NULL DEFAULT FALSE,     -- 是否上线
    is_live                  SMALLINT NOT NULL DEFAULT 0,        -- 事件状态：0=正在进行, 1=未来事件, 2=已结束
    is_sports                BOOLEAN NOT NULL DEFAULT TRUE,      -- 是否为体育事件
    stage                    VARCHAR(20) NOT NULL DEFAULT 'Q1',  -- 比赛阶段（Q1、Q2、H1、H2、HT 等）
    created_at               TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,  -- 创建时间
    updated_at               TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP  -- 更新时间
);
CREATE INDEX IF NOT EXISTS idx_event_guid ON event(guid);
CREATE INDEX IF NOT EXISTS idx_event_category_guid ON event(category_guid);  -- 新增：按分类查询
CREATE INDEX IF NOT EXISTS idx_event_ecosystem_guid ON event(ecosystem_guid);  -- 新增：按生态查询
CREATE INDEX IF NOT EXISTS idx_event_is_live ON event(is_live);  -- 新增：按状态查询
CREATE INDEX IF NOT EXISTS idx_event_is_online ON event(is_online);  -- 新增：按上线状态查询
CREATE INDEX IF NOT EXISTS idx_event_event_type ON event(event_type);  -- 新增：按排序类型查询
CREATE INDEX IF NOT EXISTS idx_event_created_at ON event(created_at);  -- 新增：按创建时间查询

-- 事件多语言表 --
CREATE TABLE IF NOT EXISTS event_language (
    guid               TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    event_guid         VARCHAR(500) NOT NULL,
    language_guid      VARCHAR(500) NOT NULL,
    title              VARCHAR(200) NOT NULL,
    rules              TEXT NOT NULL,
    created_at         TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_event_language_guid ON event_language(guid);
CREATE INDEX IF NOT EXISTS idx_event_language_event_guid ON event_language(event_guid);
CREATE INDEX IF NOT EXISTS idx_event_language_language_guid ON event_language(language_guid);

-- 事件周期表 --
CREATE TABLE IF NOT EXISTS event_result_period (
    guid               TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    event_guid         VARCHAR(500) NOT NULL DEFAULT '0',
    datetime           VARCHAR(500) NOT NULL,
    created_at         TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_event_result_period_guid ON event_result_period(guid);
CREATE INDEX IF NOT EXISTS idx_event_result_period_event_guid ON event_result_period(event_guid);

-- ============================================
-- 事件子表 (Sub Event)
-- 定义事件的具体投注选项
-- ============================================
CREATE TABLE IF NOT EXISTS sub_event(
    guid                TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),  -- 主键：子事件唯一标识
    event_guid          VARCHAR(500) NOT NULL,                   -- 父事件 GUID（外键关联 event 表）
    logo                VARCHAR(500) NOT NULL,                   -- 子事件图标 URL
    trade_volume        NUMERIC(38,18) NOT NULL DEFAULT 0,       -- 交易量
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,           -- 是否启用（新增字段）
    end_at              TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,  -- 事件结束时间
    created_at          TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,  -- 创建时间
    updated_at          TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP   -- 更新时间
);
CREATE INDEX IF NOT EXISTS idx_sub_event_guid ON sub_event(guid);
CREATE INDEX IF NOT EXISTS idx_sub_event_event_guid ON sub_event(event_guid);
CREATE INDEX IF NOT EXISTS idx_sub_event_active ON sub_event(is_active);  -- 新增：按状态查询

-- 子事件多语言表 --
CREATE TABLE IF NOT EXISTS sub_event_language (
    guid               TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    language_guid      VARCHAR(500) NOT NULL,
    sub_event_guid     VARCHAR(500) NOT NULL,
    title              VARCHAR(200) NOT NULL,
    created_at         TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_sub_event_language_guid ON sub_event_language(guid);
CREATE INDEX IF NOT EXISTS idx_sub_event_language_language_guid ON sub_event_language(language_guid);
CREATE INDEX IF NOT EXISTS idx_sub_event_language_sub_event_guid ON sub_event_language(sub_event_guid);

-- ============================================
-- 子事件方向表 (Sub Event Direction)
-- 定义子事件的投注方向和赔率
-- ============================================
CREATE TABLE IF NOT EXISTS sub_event_direction (
    guid                TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),  -- 主键：方向唯一标识
    event_guid          VARCHAR(500) NOT NULL,                   -- 父事件 GUID（外键关联 event 表）
    sub_event_guid      VARCHAR(500) NOT NULL,                   -- 子事件 GUID（外键关联 sub_event 表）
    direction           VARCHAR(200) NOT NULL DEFAULT 'Yes',     -- 其他类型：Yes、No; 运动类：具体球队
    info                JSONB NOT NULL DEFAULT '{}'::jsonb,      -- 附加信息（JSON 格式）
    is_win              BOOLEAN NOT NULL DEFAULT TRUE,           -- 是胜出
    created_at          TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,  -- 创建时间
    updated_at          TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP   -- 更新时间
);
CREATE INDEX IF NOT EXISTS idx_sub_event_direction_guid ON sub_event_direction(guid);
CREATE INDEX IF NOT EXISTS idx_sub_event_direction_sub_event_guid ON sub_event_direction(sub_event_guid);
CREATE UNIQUE INDEX IF NOT EXISTS uq_sub_event_direction ON sub_event_direction(sub_event_guid, direction);
