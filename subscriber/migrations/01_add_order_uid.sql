CREATE EXTENSION IF NOT EXISTS "uuid-message";
SET TIME ZONE 'Europe/Moscow';

CREATE TABLE orders
(
    order_uid          UUID DEFAULT uuid_generate_v4() PRIMARY KEY, -- Уникальный идентификатор заказа
    track_number       VARCHAR(50),      -- Трек-номер
    entry              VARCHAR(50),      -- Точка входа
    locale             VARCHAR(10),      -- Локаль
    internal_signature VARCHAR(255),     -- Внутренняя подпись
    customer_id        VARCHAR(50),      -- Идентификатор клиента
    delivery_service   VARCHAR(50),      -- Служба доставки
    shardkey           VARCHAR(10),      -- Ключ шардирования
    sm_id              INT,              -- Идентификатор мета-информации
    date_created       TIMESTAMP,        -- Дата создания заказа
    oof_shard          VARCHAR(10)       -- Шард OOF
);

-- Таблица delivery
CREATE TABLE delivery
(
    id        SERIAL PRIMARY KEY,                                   -- Уникальный идентификатор доставки
    order_uid UUID REFERENCES orders (order_uid) ON DELETE CASCADE, -- Связь с заказом
    name      VARCHAR(255) NOT NULL,                                -- Имя получателя
    phone     VARCHAR(20)  NOT NULL,                                -- Телефон
    zip       VARCHAR(20)  NOT NULL,                                -- Почтовый индекс
    city      VARCHAR(255) NOT NULL,                                -- Город
    address   VARCHAR(255) NOT NULL,                                -- Адрес
    region    VARCHAR(255) NOT NULL,                                -- Регион
    email     VARCHAR(255) NOT NULL                                 -- Электронная почта
);

-- Таблица payment
CREATE TABLE payment
(
    id            SERIAL PRIMARY KEY,                                   -- Уникальный идентификатор платежа
    order_uid     UUID REFERENCES orders (order_uid) ON DELETE CASCADE, -- Связь с заказом
    transaction   VARCHAR(50)    NOT NULL,                              -- Идентификатор транзакции
    request_id    VARCHAR(50),                                          -- Идентификатор запроса (может быть пустым)
    currency      VARCHAR(10)    NOT NULL,                              -- Валюта платежа
    provider      VARCHAR(50)    NOT NULL,                              -- Платежный провайдер
    amount        NUMERIC(10, 2) NOT NULL,                              -- Сумма платежа
    payment_dt    TIMESTAMP      NOT NULL,                              -- Дата и время платежа
    bank          VARCHAR(50)    NOT NULL,                              -- Банк
    delivery_cost NUMERIC(10, 2) NOT NULL,                              -- Стоимость доставки
    goods_total   NUMERIC(10, 2) NOT NULL,                              -- Общая стоимость товаров
    custom_fee    NUMERIC(10, 2) NOT NULL DEFAULT 0                     -- Таможенные сборы
);

-- Таблица items
CREATE TABLE items
(
    id           SERIAL PRIMARY KEY,                                   -- Уникальный идентификатор товара
    order_uid    UUID REFERENCES orders (order_uid) ON DELETE CASCADE, -- Связь с заказом
    chrt_id      INT            NOT NULL,                              -- Идентификатор чарта
    track_number VARCHAR(50)    NOT NULL,                              -- Трек-номер
    price        NUMERIC(10, 2) NOT NULL,                              -- Цена товара
    rid          VARCHAR(50)    NOT NULL,                              -- Идентификатор RID
    name         VARCHAR(255)   NOT NULL,                              -- Название товара
    sale         INT            NOT NULL,                              -- Скидка
    size         VARCHAR(10)    NOT NULL,                              -- Размер
    total_price  NUMERIC(10, 2) NOT NULL,                              -- Итоговая цена
    nm_id        INT            NOT NULL,                              -- Идентификатор NM
    brand        VARCHAR(255)   NOT NULL,                              -- Бренд
    status       INT            NOT NULL                               -- Статус
);

