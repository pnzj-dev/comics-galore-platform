-- NowPayments subscription statuses include a "waiting for payment" state
-- (WAITING_PAY). Widen the subscriptions.status enum so we can persist it and
-- later downgrade users whose subscription has been waiting too long.
ALTER TABLE subscriptions DROP CONSTRAINT subscriptions_status_check;
ALTER TABLE subscriptions ADD CONSTRAINT subscriptions_status_check
    CHECK (status IN ('pending', 'active', 'waiting_pay', 'expired', 'cancelled', 'failed'));
