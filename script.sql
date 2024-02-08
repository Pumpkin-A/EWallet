CREATE TABLE IF NOT EXISTS public.history (
    history_ID VARCHAR(100) NOT NULL,
    transfer_time TIMESTAMP WITH TIME ZONE NOT NULL,
    sender VARCHAR(100) NOT NULL,
    receiver VARCHAR(100) NOT NULL,
    amount FLOAT4 NOT NULL,
    CONSTRAINT history_pk PRIMARY KEY (history_ID)
);

CREATE TABLE IF NOT EXISTS public.wallet_main (
    wallet_ID VARCHAR(100) NOT NULL,
    balance FLOAT4 NOT NULL,
    CONSTRAINT wallet_pk PRIMARY KEY (wallet_ID)
);

CREATE TABLE IF NOT EXISTS public.main_and_history (
    wallet_ID_in_connection VARCHAR(100) NOT NULL,
    history_ID_in_connection VARCHAR(100) NOT NULL,
    CONSTRAINT main_and_history_pk PRIMARY KEY (wallet_ID_in_connection, history_ID_in_connection),
    CONSTRAINT history_main_and_history_fk FOREIGN KEY (history_ID_in_connection)
        REFERENCES public.history (history_ID) ON DELETE NO ACTION ON UPDATE NO ACTION,
    CONSTRAINT wallet_main_main_and_history_fk FOREIGN KEY (wallet_ID_in_connection)
        REFERENCES public.wallet_main (wallet_ID) ON DELETE NO ACTION ON UPDATE NO ACTION
);