CREATE SEQUENCE customerorder_id_seq;

CREATE TABLE userinfo (
	cellnumber varchar(15) NOT NULL,
	nickname varchar(255) NULL,
	email varchar(255) NULL,
	socialmedia varchar(255) NULL,
	consent bool NULL,
	datetimejoined timestamp NULL,
	CONSTRAINT userinfo_pkey PRIMARY KEY (cellnumber)
);

CREATE TYPE pricingTypeEnum AS ENUM ('WeightItem', 'SingleItem');

CREATE TABLE catalogueitem (
	catalogueID varchar(255) NOT NULL,
	catalogueitemID int NOT NULL,
	"selection" varchar(255) NULL,
	"item" varchar(255) NULL,
	"options" varchar(255) NULL,
	pricingType pricingTypeEnum,
	CONSTRAINT catalogueitem_pk PRIMARY KEY (catalogueID, catalogueitemID)
);

CREATE TABLE customerorder (
    orderID int PRIMARY KEY DEFAULT nextval('customerorder_id_seq'),
    cellnumber varchar(15) REFERENCES userinfo(cellnumber),
    catalogueID varchar(255) NOT NULL,
    orderitems varchar(255) NULL,
    orderTotal numeric(12,0),
    ispaid bool NULL,
    datetimedelivered timestamp NULL,
    isclosed bool NULL
);
