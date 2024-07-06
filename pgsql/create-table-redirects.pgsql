-- Table: public.redirects

-- DROP TABLE IF EXISTS public.redirects;

CREATE TABLE IF NOT EXISTS public.redirects
(
    id integer NOT NULL DEFAULT nextval('redirects_id_seq'::regclass),
    host character varying COLLATE pg_catalog."default" NOT NULL,
    address character varying COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT redirects_pkey PRIMARY KEY (id),
    CONSTRAINT redirects_host_key UNIQUE (host)
)

TABLESPACE pg_default;

ALTER TABLE IF EXISTS public.redirects
    OWNER to postgres;
    