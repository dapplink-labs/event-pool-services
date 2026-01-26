DO
$$
   BEGIN
        IF
            NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'uint256') THEN
            CREATE DOMAIN UINT256 AS NUMERIC CHECK (VALUE >= 0 AND VALUE < POWER(CAST(2 AS NUMERIC), CAST(256 AS NUMERIC)) AND SCALE(VALUE) = 0);
        ELSE
            ALTER DOMAIN UINT256 DROP CONSTRAINT uint256_check;
            ALTER DOMAIN UINT256 ADD CHECK (VALUE >= 0 AND VALUE < POWER(CAST(2 AS NUMERIC), CAST(256 AS NUMERIC)) AND SCALE(VALUE) = 0);
        END IF;
   END
$$;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp" cascade;

create table if not exists sys_log (
    guid            TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    action          VARCHAR(100) DEFAULT '', -- 路径 --
    "desc"          VARCHAR(100) DEFAULT '', -- 描述 --
    admin           VARCHAR(30)  DEFAULT '', -- 操作管理员 --
    ip              VARCHAR(30)  DEFAULT '', -- 操作管理员 IP --
    cate            SMALLINT DEFAULT 0,      -- 类型(0表示其他;1=>表示登陆;2=>表示财务操作) --
    status          SMALLINT DEFAULT -1,     -- 登陆状态(0=>成功;1=>失败) --
    asset           VARCHAR(255) DEFAULT '', -- 币种 --
    before          VARCHAR(255) DEFAULT '', -- 修改前 --
    after           VARCHAR(255) DEFAULT '', -- 修改后 --
    user_id         BIGINT DEFAULT 0,
    order_number    VARCHAR(64) DEFAULT '',
    op              SMALLINT DEFAULT -1,     -- 操作类型(0添加;1编辑)
    created_at      TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_syslog_cate ON sys_log (cate);
CREATE INDEX idx_syslog_status ON sys_log (status);
CREATE INDEX idx_syslog_order_number ON sys_log (order_number);


create table if not exists auth (
    guid            TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    auth_name       VARCHAR(255) DEFAULT '', -- 权限名称
    auth_url        VARCHAR(255) DEFAULT '', -- 权限路径/接口地址
    user_id         INT DEFAULT 0,           -- 所属用户/管理员ID
    pid             INT DEFAULT 0,           -- 父级权限ID
    sort            INT DEFAULT 0,           -- 排序
    icon            VARCHAR(255) DEFAULT '', -- 图标
    is_show         INT DEFAULT 1,           -- 是否显示(1显示;0隐藏)
    status          INT DEFAULT 1,           -- 状态(1启用;0禁用)
    create_id       INT DEFAULT 0,           -- 创建人ID
    update_id       INT DEFAULT 0,           -- 修改人ID
    created_at      TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_auth_user_id   ON auth (user_id);
CREATE INDEX idx_auth_pid       ON auth (pid);
CREATE INDEX idx_auth_create_id ON auth (create_id);
CREATE INDEX idx_auth_update_id ON auth (update_id);

create table if not exists role (
    guid         TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    role_name    VARCHAR(100) DEFAULT '', -- 角色名称
    detail       VARCHAR(255) DEFAULT '', -- 角色描述/说明
    status       INT DEFAULT 1, -- 状态(1启用;0禁用)
    create_id    INT DEFAULT 0, -- 创建人ID
    update_id    INT DEFAULT 0, -- 修改人ID
    created_at   TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_role_role_name ON role (role_name);

create table if not exists role_auth (
    auth_id    INT NOT NULL,
    role_id    BIGINT NOT NULL,
    PRIMARY KEY (auth_id, role_id)
);
CREATE INDEX idx_role_auth_role_id ON role_auth (role_id);

create table if not exists admin (
    guid          TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    login_name    VARCHAR(32)  NOT NULL UNIQUE,   -- 登录名 --
    real_name     VARCHAR(32)  UNIQUE,            -- 真实姓名 --
    password      VARCHAR(100) NOT NULL,          -- 密码(加密后) --
    role_ids      VARCHAR(255) DEFAULT '',        -- 角色 ID 列表（字符串存 JSON/CSV） --
    phone         VARCHAR(11) UNIQUE,             -- 手机号 --
    email         VARCHAR(32),                    -- 邮箱 --
    salt          VARCHAR(255) DEFAULT '',        -- 密码盐 --
    last_login    BIGINT DEFAULT 0,               -- 最后登录时间戳 --
    last_ip       VARCHAR(255) DEFAULT '',        -- 最后登录 IP --
    status        INT DEFAULT 1,                  -- 状态(1启用;0禁用) --
    create_id     INT DEFAULT 0,                  -- 创建人 --
    update_id     INT DEFAULT 0,                  -- 修改人 --
    created_at    TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_admin_status ON admin (status);
CREATE INDEX idx_admin_create_id ON admin (create_id);
CREATE INDEX idx_admin_update_id ON admin (update_id);
CREATE INDEX idx_admin_last_login ON admin (last_login);

-- 支持的语言表 --
CREATE TABLE IF NOT EXISTS languages (
    guid             TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    language_name    VARCHAR DEFAULT 'zh',
    created_at       TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_languages ON languages (guid);




-- 事件分类表 --
CREATE TABLE IF NOT EXISTS category (
    guid            TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_category_guid ON category(guid);

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

-- 所属生态 --
CREATE TABLE IF NOT EXISTS ecosystem (
    guid            TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    category_guid   VARCHAR(255) NOT NULL,
    event_num       UINT256 NOT NULL CHECK (event_num > 0),  -- 生态的事件个数 --
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_ecosystem_guid ON ecosystem(guid);

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

-- 时间标签表 --
CREATE TABLE IF NOT EXISTS event_period (
    guid              TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    category_guid     VARCHAR(255) NOT NULL,
    is_active         BOOLEAN NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);

-- 时间标签表多语言表 --
CREATE TABLE IF NOT EXISTS event_period_language (
    guid                TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    event_period_guid   VARCHAR(255)  NOT NULL,
    language_guid       VARCHAR(255)  NOT NULL,
    name                VARCHAR(50) NOT NULL,
    created_at          TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);

-- 运动类团队 --
CREATE TABLE IF NOT EXISTS team_group (
    guid                TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    logo                VARCHAR(255)  NOT NULL,
    created_at          TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS team_group_language (
    guid                TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    team_group_guid     VARCHAR(255)  NOT NULL,
    name                VARCHAR(255)  NOT NULL,
    created_at          TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);

-- 事件表 --
CREATE TABLE IF NOT EXISTS event (
    guid                     TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    category_guid            VARCHAR(500) NOT NULL,
    ecosystem_guid           VARCHAR(500) NOT NULL,
    event_period_guid        VARCHAR(500) NOT NULL,
    main_team_group_guid     VARCHAR(255) NOT NULL, -- 非运动类的 ID 为 0 --
    cluster_team_group_guid  VARCHAR(255) NOT NULL, -- 非运动类的 ID 为 0 --
    main_score               UINT256 NOT NULL,
    cluster_score            UINT256 NOT NULL,
    logo                     VARCHAR(300) NOT NULL,
    order_type               SMALLINT NOT NULL DEFAULT '0', -- 0:热门话题; 1:突发; 2:最新--
    order_num                UINT256 NOT NULL,              -- 订单数量 --
    open_time                VARCHAR(100) NOT NULL,
    trade_volume             NUMERIC(32,16) NOT NULL DEFAULT 0,
    experiment_result        TEXT NOT NULL,
    info                     JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_online                BOOLEAN NOT NULL DEFAULT FALSE,
    is_live                  SMALLINT NOT NULL DEFAULT 0, -- 0:正在发送事件; 1:未来事件；2:已结束事件--
    is_sports                BOOLEAN NOT NULL DEFAULT TRUE,
    stage                    VARCHAR(20) NOT NULL DEFAULT 'Q1', -- Q1,Q2,H1,H2,HT 等--
    created_at               TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at               TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);

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

-- 事件子表 --
CREATE TABLE IF NOT EXISTS sub_event(
    guid                TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    parent_event_guid   VARCHAR(500) NOT NULL,
    title               VARCHAR(200) NOT NULL,
    logo                VARCHAR(300) NOT NULL,
    trade_volume        NUMERIC(32,16) NOT NULL DEFAULT 0,
    created_at          TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_sub_event_parent_event_guid ON sub_event(parent_event_guid);

-- 子事件多语言表 --
CREATE TABLE IF NOT EXISTS sub_event_language (
    guid               TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    language_guid      VARCHAR(500) NOT NULL,
    sub_event_guid     VARCHAR(500) NOT NULL,
    title              VARCHAR(200) NOT NULL,
    created_at         TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);

-- 子事件方向表 --
CREATE TABLE IF NOT EXISTS sub_event_direction (
    guid                TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    sub_event_guid      VARCHAR(500) NOT NULL,
    direction           VARCHAR(200) NOT NULL DEFAULT 'Yes', -- Yes, No, 球塞比分 --
    chance              SMALLINT NOT NULL,
    new_ask_price       UINT256 NOT NULL DEFAULT '0',
    new_bid_price       UINT256 NOT NULL DEFAULT '0',
    info                JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at          TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_sub_event_direction_sub_event_guid ON sub_event_direction(sub_event_guid);

-- 子事件方向表 --
CREATE TABLE IF NOT EXISTS sub_event_chance_stat (
    guid                TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    sub_event_guid      VARCHAR(500) NOT NULL,
    chance              SMALLINT NOT NULL,
    datetime            VARCHAR(500) NOT NULL,
    stat_way            SMALLINT NOT NULL, --0:1h; 2:6h; 3:1d; 4:1w; 5:All--
    created_at         TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_sub_event_chance_stat_sub_event_guid ON sub_event_chance_stat(sub_event_guid);




