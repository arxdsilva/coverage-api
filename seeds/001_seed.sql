-- DEV SEED DATA
-- Remove this file and the compose seed service before prod deployment.

INSERT INTO projects (id, project_key, name, group_name, default_branch, global_threshold_percent)
VALUES
  ('11111111-1111-1111-1111-111111111111', 'acme/coverage-api', 'coverage-api', 'infrastructure/internal', 'main', 80.00),
  ('22222222-2222-2222-2222-222222222222', 'acme/payments', 'payments', 'payments', 'main', 85.00),
  ('33333333-3333-3333-3333-333333333333', 'acme/webapp', 'webapp', 'frontend', 'main', 75.00),
  ('44444444-4444-4444-4444-444444444444', 'acme/auth-service', 'auth-service', 'auth', 'main', 82.50),
  ('55555555-5555-5555-5555-555555555555', 'acme/notifications', 'notifications', 'notifications', 'main', 78.00),
  ('66666666-6666-6666-6666-666666666666', 'acme/catalog', 'catalog', 'catalog', 'main', 81.00),
  ('77777777-7777-7777-7777-777777777777', 'acme/orders', 'orders', 'orders', 'main', 84.00),
  ('88888888-8888-8888-8888-888888888888', 'acme/shipping', 'shipping', 'fulfillment', 'main', 79.50),
  ('99999999-9999-9999-9999-999999999999', 'acme/analytics', 'analytics', 'analytics', 'main', 77.00),
  ('aaaaaaaa-1111-1111-1111-111111111111', 'acme/search', 'search', 'catalog', 'main', 83.00)
ON CONFLICT (id) DO NOTHING;

INSERT INTO coverage_runs (id, project_id, branch, commit_sha, author, trigger_type, run_timestamp, total_coverage_percent)
VALUES
  ('10000000-0000-0000-0000-000000000001', '11111111-1111-1111-1111-111111111111', 'main', 'ca11aa01', 'dev-seed', 'push', NOW() - INTERVAL '8 day', 80.10),
  ('10000000-0000-0000-0000-000000000002', '11111111-1111-1111-1111-111111111111', 'develop', 'ca11aa02', 'dev-seed', 'push', NOW() - INTERVAL '7 day', 79.40),
  ('10000000-0000-0000-0000-000000000003', '11111111-1111-1111-1111-111111111111', 'main', 'ca11aa03', 'dev-seed', 'push', NOW() - INTERVAL '2 day', 83.40),
  ('10000000-0000-0000-0000-000000000004', '11111111-1111-1111-1111-111111111111', 'feature/rbac', 'ca11aa04', 'dev-seed', 'pr', NOW() - INTERVAL '1 day', 82.80),

  ('20000000-0000-0000-0000-000000000001', '22222222-2222-2222-2222-222222222222', 'main', 'pay10001', 'dev-seed', 'push', NOW() - INTERVAL '9 day', 86.90),
  ('20000000-0000-0000-0000-000000000002', '22222222-2222-2222-2222-222222222222', 'release/1.4', 'pay10002', 'dev-seed', 'manual', NOW() - INTERVAL '6 day', 85.20),
  ('20000000-0000-0000-0000-000000000003', '22222222-2222-2222-2222-222222222222', 'main', 'pay10003', 'dev-seed', 'push', NOW() - INTERVAL '2 day', 84.70),
  ('20000000-0000-0000-0000-000000000004', '22222222-2222-2222-2222-222222222222', 'feature/fraud-check', 'pay10004', 'dev-seed', 'pr', NOW() - INTERVAL '1 day', 83.60),

  ('30000000-0000-0000-0000-000000000001', '33333333-3333-3333-3333-333333333333', 'main', 'web10001', 'dev-seed', 'push', NOW() - INTERVAL '10 day', 71.80),
  ('30000000-0000-0000-0000-000000000002', '33333333-3333-3333-3333-333333333333', 'develop', 'web10002', 'dev-seed', 'push', NOW() - INTERVAL '7 day', 73.10),
  ('30000000-0000-0000-0000-000000000003', '33333333-3333-3333-3333-333333333333', 'main', 'web10003', 'dev-seed', 'push', NOW() - INTERVAL '2 day', 76.50),
  ('30000000-0000-0000-0000-000000000004', '33333333-3333-3333-3333-333333333333', 'feature/ui-refresh', 'web10004', 'dev-seed', 'pr', NOW() - INTERVAL '20 hour', 77.10),

  ('40000000-0000-0000-0000-000000000001', '44444444-4444-4444-4444-444444444444', 'main', 'aut10001', 'dev-seed', 'push', NOW() - INTERVAL '11 day', 81.40),
  ('40000000-0000-0000-0000-000000000002', '44444444-4444-4444-4444-444444444444', 'develop', 'aut10002', 'dev-seed', 'push', NOW() - INTERVAL '6 day', 82.20),
  ('40000000-0000-0000-0000-000000000003', '44444444-4444-4444-4444-444444444444', 'main', 'aut10003', 'dev-seed', 'push', NOW() - INTERVAL '2 day', 83.00),
  ('40000000-0000-0000-0000-000000000004', '44444444-4444-4444-4444-444444444444', 'feature/oidc-claims', 'aut10004', 'dev-seed', 'pr', NOW() - INTERVAL '22 hour', 82.10),

  ('50000000-0000-0000-0000-000000000001', '55555555-5555-5555-5555-555555555555', 'main', 'not10001', 'dev-seed', 'push', NOW() - INTERVAL '9 day', 76.30),
  ('50000000-0000-0000-0000-000000000002', '55555555-5555-5555-5555-555555555555', 'develop', 'not10002', 'dev-seed', 'push', NOW() - INTERVAL '6 day', 77.20),
  ('50000000-0000-0000-0000-000000000003', '55555555-5555-5555-5555-555555555555', 'main', 'not10003', 'dev-seed', 'push', NOW() - INTERVAL '2 day', 78.60),
  ('50000000-0000-0000-0000-000000000004', '55555555-5555-5555-5555-555555555555', 'feature/batch-retry', 'not10004', 'dev-seed', 'pr', NOW() - INTERVAL '18 hour', 79.10),

  ('60000000-0000-0000-0000-000000000001', '66666666-6666-6666-6666-666666666666', 'main', 'cat10001', 'dev-seed', 'push', NOW() - INTERVAL '8 day', 80.50),
  ('60000000-0000-0000-0000-000000000002', '66666666-6666-6666-6666-666666666666', 'release/2.2', 'cat10002', 'dev-seed', 'manual', NOW() - INTERVAL '5 day', 81.30),
  ('60000000-0000-0000-0000-000000000003', '66666666-6666-6666-6666-666666666666', 'main', 'cat10003', 'dev-seed', 'push', NOW() - INTERVAL '2 day', 82.60),
  ('60000000-0000-0000-0000-000000000004', '66666666-6666-6666-6666-666666666666', 'feature/stock-sync', 'cat10004', 'dev-seed', 'pr', NOW() - INTERVAL '16 hour', 83.10),

  ('70000000-0000-0000-0000-000000000001', '77777777-7777-7777-7777-777777777777', 'main', 'ord10001', 'dev-seed', 'push', NOW() - INTERVAL '9 day', 83.10),
  ('70000000-0000-0000-0000-000000000002', '77777777-7777-7777-7777-777777777777', 'develop', 'ord10002', 'dev-seed', 'push', NOW() - INTERVAL '6 day', 84.00),
  ('70000000-0000-0000-0000-000000000003', '77777777-7777-7777-7777-777777777777', 'main', 'ord10003', 'dev-seed', 'push', NOW() - INTERVAL '2 day', 85.30),
  ('70000000-0000-0000-0000-000000000004', '77777777-7777-7777-7777-777777777777', 'feature/partial-cancel', 'ord10004', 'dev-seed', 'pr', NOW() - INTERVAL '14 hour', 84.10),

  ('80000000-0000-0000-0000-000000000001', '88888888-8888-8888-8888-888888888888', 'main', 'shp10001', 'dev-seed', 'push', NOW() - INTERVAL '10 day', 78.20),
  ('80000000-0000-0000-0000-000000000002', '88888888-8888-8888-8888-888888888888', 'develop', 'shp10002', 'dev-seed', 'push', NOW() - INTERVAL '5 day', 79.00),
  ('80000000-0000-0000-0000-000000000003', '88888888-8888-8888-8888-888888888888', 'main', 'shp10003', 'dev-seed', 'push', NOW() - INTERVAL '2 day', 80.10),
  ('80000000-0000-0000-0000-000000000004', '88888888-8888-8888-8888-888888888888', 'feature/label-cache', 'shp10004', 'dev-seed', 'pr', NOW() - INTERVAL '12 hour', 79.60),

  ('90000000-0000-0000-0000-000000000001', '99999999-9999-9999-9999-999999999999', 'main', 'ana10001', 'dev-seed', 'push', NOW() - INTERVAL '12 day', 75.20),
  ('90000000-0000-0000-0000-000000000002', '99999999-9999-9999-9999-999999999999', 'develop', 'ana10002', 'dev-seed', 'push', NOW() - INTERVAL '7 day', 76.00),
  ('90000000-0000-0000-0000-000000000003', '99999999-9999-9999-9999-999999999999', 'main', 'ana10003', 'dev-seed', 'push', NOW() - INTERVAL '2 day', 77.80),
  ('90000000-0000-0000-0000-000000000004', '99999999-9999-9999-9999-999999999999', 'feature/dashboard-v2', 'ana10004', 'dev-seed', 'pr', NOW() - INTERVAL '10 hour', 78.50),

  ('a0000000-0000-0000-0000-000000000001', 'aaaaaaaa-1111-1111-1111-111111111111', 'main', 'sea10001', 'dev-seed', 'push', NOW() - INTERVAL '8 day', 82.10),
  ('a0000000-0000-0000-0000-000000000002', 'aaaaaaaa-1111-1111-1111-111111111111', 'release/3.0', 'sea10002', 'dev-seed', 'manual', NOW() - INTERVAL '5 day', 83.00),
  ('a0000000-0000-0000-0000-000000000003', 'aaaaaaaa-1111-1111-1111-111111111111', 'main', 'sea10003', 'dev-seed', 'push', NOW() - INTERVAL '2 day', 84.20),
  ('a0000000-0000-0000-0000-000000000004', 'aaaaaaaa-1111-1111-1111-111111111111', 'feature/relevance-tuning', 'sea10004', 'dev-seed', 'pr', NOW() - INTERVAL '8 hour', 83.50)
ON CONFLICT (id) DO NOTHING;

INSERT INTO package_coverages (id, run_id, package_import_path, coverage_percent)
VALUES
  ('b1000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000003', 'github.com/acme/coverage-api/internal/adapters/http', 82.80),
  ('b1000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000003', 'github.com/acme/coverage-api/internal/application', 84.00),
  ('b1000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000004', 'github.com/acme/coverage-api/internal/adapters/auth', 81.90),
  ('b1000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000004', 'github.com/acme/coverage-api/internal/adapters/postgres', 83.30),

  ('b2000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000003', 'github.com/acme/payments/internal/service', 83.90),
  ('b2000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000003', 'github.com/acme/payments/internal/api', 86.10),
  ('b2000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000004', 'github.com/acme/payments/internal/fraud', 82.70),
  ('b2000000-0000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000004', 'github.com/acme/payments/internal/checkout', 84.20),

  ('b3000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000003', 'github.com/acme/webapp/internal/web', 74.20),
  ('b3000000-0000-0000-0000-000000000002', '30000000-0000-0000-0000-000000000003', 'github.com/acme/webapp/internal/core', 78.30),
  ('b3000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000004', 'github.com/acme/webapp/internal/layout', 76.80),
  ('b3000000-0000-0000-0000-000000000004', '30000000-0000-0000-0000-000000000004', 'github.com/acme/webapp/internal/components', 77.40),

  ('b4000000-0000-0000-0000-000000000001', '40000000-0000-0000-0000-000000000003', 'github.com/acme/auth-service/internal/tokens', 83.40),
  ('b4000000-0000-0000-0000-000000000002', '40000000-0000-0000-0000-000000000003', 'github.com/acme/auth-service/internal/oidc', 82.60),
  ('b5000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000003', 'github.com/acme/notifications/internal/email', 78.10),
  ('b5000000-0000-0000-0000-000000000002', '50000000-0000-0000-0000-000000000004', 'github.com/acme/notifications/internal/retry', 79.30),

  ('b6000000-0000-0000-0000-000000000001', '60000000-0000-0000-0000-000000000003', 'github.com/acme/catalog/internal/ingest', 82.70),
  ('b6000000-0000-0000-0000-000000000002', '60000000-0000-0000-0000-000000000004', 'github.com/acme/catalog/internal/search', 83.20),
  ('b7000000-0000-0000-0000-000000000001', '70000000-0000-0000-0000-000000000003', 'github.com/acme/orders/internal/workflow', 85.60),
  ('b7000000-0000-0000-0000-000000000002', '70000000-0000-0000-0000-000000000004', 'github.com/acme/orders/internal/refund', 83.80),

  ('b8000000-0000-0000-0000-000000000001', '80000000-0000-0000-0000-000000000003', 'github.com/acme/shipping/internal/labels', 80.00),
  ('b8000000-0000-0000-0000-000000000002', '80000000-0000-0000-0000-000000000004', 'github.com/acme/shipping/internal/rates', 79.40),
  ('b9000000-0000-0000-0000-000000000001', '90000000-0000-0000-0000-000000000003', 'github.com/acme/analytics/internal/etl', 77.10),
  ('b9000000-0000-0000-0000-000000000002', '90000000-0000-0000-0000-000000000004', 'github.com/acme/analytics/internal/dashboards', 78.70),

  ('ba000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000003', 'github.com/acme/search/internal/index', 84.00),
  ('ba000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000004', 'github.com/acme/search/internal/ranking', 83.60)
ON CONFLICT (id) DO NOTHING;

INSERT INTO integration_test_runs (
  id, project_id, branch, commit_sha, author, trigger_type, run_timestamp,
  ginkgo_version, suite_description, suite_path, total_specs, passed_specs,
  failed_specs, skipped_specs, flaked_specs, pending_specs, interrupted,
  timed_out, duration_ms, status, environment
)
VALUES
  ('c1000000-0000-0000-0000-000000000001', '11111111-1111-1111-1111-111111111111', 'main', 'ca11it01', 'dev-seed', 'push', NOW() - INTERVAL '2 day', '2.21.0', 'coverage-api integration suite', './integration', 12, 11, 1, 0, 0, 0, FALSE, FALSE, 18203, 'failed', 'test'),
  ('c1000000-0000-0000-0000-000000000002', '11111111-1111-1111-1111-111111111111', 'feature/rbac', 'ca11it02', 'dev-seed', 'pr', NOW() - INTERVAL '20 hour', '2.21.0', 'coverage-api integration suite', './integration', 10, 10, 0, 0, 0, 0, FALSE, FALSE, 14110, 'passed', 'stage'),

  ('c2000000-0000-0000-0000-000000000001', '22222222-2222-2222-2222-222222222222', 'main', 'payit001', 'dev-seed', 'push', NOW() - INTERVAL '2 day', '2.21.0', 'payments integration suite', './integration', 25, 24, 0, 1, 0, 0, FALSE, FALSE, 26350, 'passed', 'prod'),
  ('c2000000-0000-0000-0000-000000000002', '22222222-2222-2222-2222-222222222222', 'feature/fraud-check', 'payit002', 'dev-seed', 'pr', NOW() - INTERVAL '16 hour', '2.21.0', 'payments integration suite', './integration', 18, 16, 1, 0, 1, 0, FALSE, FALSE, 21900, 'failed', 'test'),

  ('c3000000-0000-0000-0000-000000000001', '33333333-3333-3333-3333-333333333333', 'main', 'webit001', 'dev-seed', 'push', NOW() - INTERVAL '30 hour', '2.21.0', 'webapp integration suite', './integration', 20, 18, 1, 0, 0, 1, FALSE, FALSE, 24520, 'failed', 'stage'),
  ('c7000000-0000-0000-0000-000000000001', '77777777-7777-7777-7777-777777777777', 'main', 'ordit001', 'dev-seed', 'push', NOW() - INTERVAL '12 hour', '2.21.0', 'orders integration suite', './integration', 14, 14, 0, 0, 0, 0, FALSE, FALSE, 17640, 'passed', 'prod')
ON CONFLICT (id) DO NOTHING;

INSERT INTO integration_spec_results (
  id, integration_run_id, spec_path, leaf_node_text, state, duration_ms,
  failure_message, failure_location_file, failure_location_line
)
VALUES
  ('d1000000-0000-0000-0000-000000000001', 'c1000000-0000-0000-0000-000000000001', 'POST /v1/coverage-runs > rejects invalid api key', 'rejects invalid api key', 'failed', 53, 'expected 401, got 500', 'integration/auth_test.go', 84),
  ('d1000000-0000-0000-0000-000000000002', 'c1000000-0000-0000-0000-000000000001', 'POST /v1/coverage-runs > creates project on first ingest', 'creates project on first ingest', 'passed', 112, NULL, NULL, NULL),
  ('d1000000-0000-0000-0000-000000000003', 'c1000000-0000-0000-0000-000000000002', 'RBAC > denies missing project access', 'denies missing project access', 'passed', 74, NULL, NULL, NULL),
  ('d1000000-0000-0000-0000-000000000004', 'c1000000-0000-0000-0000-000000000002', 'RBAC > allows maintainer role', 'allows maintainer role', 'passed', 81, NULL, NULL, NULL),

  ('d2000000-0000-0000-0000-000000000001', 'c2000000-0000-0000-0000-000000000001', 'Checkout > creates authorized payment intent', 'creates authorized payment intent', 'passed', 120, NULL, NULL, NULL),
  ('d2000000-0000-0000-0000-000000000002', 'c2000000-0000-0000-0000-000000000001', 'Webhooks > ignores duplicate provider event', 'ignores duplicate provider event', 'skipped', 0, NULL, NULL, NULL),
  ('d2000000-0000-0000-0000-000000000003', 'c2000000-0000-0000-0000-000000000002', 'Fraud > blocks card on rule hit', 'blocks card on rule hit', 'failed', 96, 'expected status blocked, got review', 'integration/fraud_test.go', 142),
  ('d2000000-0000-0000-0000-000000000004', 'c2000000-0000-0000-0000-000000000002', 'Fraud > handles provider timeout retry', 'handles provider timeout retry', 'flaky', 203, 'passed on retry after transient timeout', 'integration/fraud_retry_test.go', 57),

  ('d3000000-0000-0000-0000-000000000001', 'c3000000-0000-0000-0000-000000000001', 'SSR > renders dashboard overview', 'renders dashboard overview', 'passed', 128, NULL, NULL, NULL),
  ('d3000000-0000-0000-0000-000000000002', 'c3000000-0000-0000-0000-000000000001', 'SSR > handles stale session token', 'handles stale session token', 'failed', 88, 'expected redirect to /login', 'integration/session_test.go', 73),
  ('d3000000-0000-0000-0000-000000000003', 'c3000000-0000-0000-0000-000000000001', 'UI > progressive hydration fallback', 'progressive hydration fallback', 'pending', 0, NULL, NULL, NULL),

  ('d7000000-0000-0000-0000-000000000001', 'c7000000-0000-0000-0000-000000000001', 'Orders > creates and reserves stock', 'creates and reserves stock', 'passed', 109, NULL, NULL, NULL),
  ('d7000000-0000-0000-0000-000000000002', 'c7000000-0000-0000-0000-000000000001', 'Orders > cancels reservation on payment failure', 'cancels reservation on payment failure', 'passed', 117, NULL, NULL, NULL)
ON CONFLICT (id) DO NOTHING;

INSERT INTO integration_test_runs (
  id, project_id, branch, commit_sha, author, trigger_type, run_timestamp,
  ginkgo_version, suite_description, suite_path, total_specs, passed_specs,
  failed_specs, skipped_specs, flaked_specs, pending_specs, interrupted,
  timed_out, duration_ms, status, environment
)
VALUES
  ('c4000000-0000-0000-0000-000000000001', '44444444-4444-4444-4444-444444444444', 'main', 'autit001', 'dev-seed', 'push', NOW() - INTERVAL '15 hour', '2.21.0', 'auth-service integration suite', './integration', 22, 21, 1, 0, 0, 0, FALSE, FALSE, 22840, 'failed', 'test'),
  ('c4000000-0000-0000-0000-000000000002', '44444444-4444-4444-4444-444444444444', 'feature/oidc-claims', 'autit002', 'dev-seed', 'pr', NOW() - INTERVAL '9 hour', '2.21.0', 'auth-service integration suite', './integration', 19, 18, 0, 1, 0, 0, FALSE, FALSE, 20120, 'passed', 'stage'),

  ('c5000000-0000-0000-0000-000000000001', '55555555-5555-5555-5555-555555555555', 'main', 'notit001', 'dev-seed', 'push', NOW() - INTERVAL '14 hour', '2.21.0', 'notifications integration suite', './integration', 16, 14, 1, 0, 1, 0, FALSE, FALSE, 19010, 'failed', 'prod'),
  ('c5000000-0000-0000-0000-000000000002', '55555555-5555-5555-5555-555555555555', 'feature/batch-retry', 'notit002', 'dev-seed', 'pr', NOW() - INTERVAL '7 hour', '2.21.0', 'notifications integration suite', './integration', 13, 13, 0, 0, 0, 0, FALSE, FALSE, 15300, 'passed', 'test'),

  ('c6000000-0000-0000-0000-000000000001', '66666666-6666-6666-6666-666666666666', 'main', 'catit001', 'dev-seed', 'push', NOW() - INTERVAL '13 hour', '2.21.0', 'catalog integration suite', './integration', 24, 22, 1, 0, 0, 1, FALSE, FALSE, 27140, 'failed', 'stage'),
  ('c6000000-0000-0000-0000-000000000002', '66666666-6666-6666-6666-666666666666', 'feature/stock-sync', 'catit002', 'dev-seed', 'pr', NOW() - INTERVAL '6 hour', '2.21.0', 'catalog integration suite', './integration', 20, 20, 0, 0, 0, 0, FALSE, FALSE, 23600, 'passed', 'prod'),

  ('c8000000-0000-0000-0000-000000000001', '88888888-8888-8888-8888-888888888888', 'main', 'shpit001', 'dev-seed', 'push', NOW() - INTERVAL '11 hour', '2.21.0', 'shipping integration suite', './integration', 15, 14, 1, 0, 0, 0, FALSE, FALSE, 18220, 'failed', 'test'),
  ('c8000000-0000-0000-0000-000000000002', '88888888-8888-8888-8888-888888888888', 'feature/label-cache', 'shpit002', 'dev-seed', 'pr', NOW() - INTERVAL '5 hour', '2.21.0', 'shipping integration suite', './integration', 12, 12, 0, 0, 0, 0, FALSE, FALSE, 14990, 'passed', 'stage'),

  ('c9000000-0000-0000-0000-000000000001', '99999999-9999-9999-9999-999999999999', 'main', 'anait001', 'dev-seed', 'push', NOW() - INTERVAL '10 hour', '2.21.0', 'analytics integration suite', './integration', 17, 15, 1, 0, 0, 1, FALSE, FALSE, 21110, 'failed', 'prod'),
  ('ca000000-0000-0000-0000-000000000001', 'aaaaaaaa-1111-1111-1111-111111111111', 'main', 'seait001', 'dev-seed', 'push', NOW() - INTERVAL '4 hour', '2.21.0', 'search integration suite', './integration', 21, 21, 0, 0, 0, 0, FALSE, FALSE, 24480, 'passed', 'test')
ON CONFLICT (id) DO NOTHING;

-- Extra main-branch integration history so each project has at least 5 main runs.
INSERT INTO integration_test_runs (
  id, project_id, branch, commit_sha, author, trigger_type, run_timestamp,
  ginkgo_version, suite_description, suite_path, total_specs, passed_specs,
  failed_specs, skipped_specs, flaked_specs, pending_specs, interrupted,
  timed_out, duration_ms, status, environment
)
VALUES
  ('e1110000-0000-0000-0000-000000000001', '11111111-1111-1111-1111-111111111111', 'main', 'ca11m005', 'dev-seed', 'push', NOW() - INTERVAL '6 day', '2.21.0', 'coverage-api integration suite', './integration', 12, 12, 0, 0, 0, 0, FALSE, FALSE, 17100, 'passed', 'stage'),
  ('e1110000-0000-0000-0000-000000000002', '11111111-1111-1111-1111-111111111111', 'main', 'ca11m006', 'dev-seed', 'push', NOW() - INTERVAL '5 day', '2.21.0', 'coverage-api integration suite', './integration', 12, 11, 1, 0, 0, 0, FALSE, FALSE, 18340, 'failed', 'prod'),
  ('e1110000-0000-0000-0000-000000000003', '11111111-1111-1111-1111-111111111111', 'main', 'ca11m007', 'dev-seed', 'push', NOW() - INTERVAL '4 day', '2.21.0', 'coverage-api integration suite', './integration', 12, 12, 0, 0, 0, 0, FALSE, FALSE, 16980, 'passed', 'test'),
  ('e1110000-0000-0000-0000-000000000004', '11111111-1111-1111-1111-111111111111', 'main', 'ca11m008', 'dev-seed', 'push', NOW() - INTERVAL '3 day', '2.21.0', 'coverage-api integration suite', './integration', 12, 11, 1, 0, 0, 0, FALSE, FALSE, 18620, 'failed', 'stage'),

  ('e2220000-0000-0000-0000-000000000001', '22222222-2222-2222-2222-222222222222', 'main', 'paym0005', 'dev-seed', 'push', NOW() - INTERVAL '6 day', '2.21.0', 'payments integration suite', './integration', 25, 25, 0, 0, 0, 0, FALSE, FALSE, 24910, 'passed', 'prod'),
  ('e2220000-0000-0000-0000-000000000002', '22222222-2222-2222-2222-222222222222', 'main', 'paym0006', 'dev-seed', 'push', NOW() - INTERVAL '5 day', '2.21.0', 'payments integration suite', './integration', 25, 24, 1, 0, 0, 0, FALSE, FALSE, 26380, 'failed', 'test'),
  ('e2220000-0000-0000-0000-000000000003', '22222222-2222-2222-2222-222222222222', 'main', 'paym0007', 'dev-seed', 'push', NOW() - INTERVAL '4 day', '2.21.0', 'payments integration suite', './integration', 25, 25, 0, 0, 0, 0, FALSE, FALSE, 24670, 'passed', 'stage'),
  ('e2220000-0000-0000-0000-000000000004', '22222222-2222-2222-2222-222222222222', 'main', 'paym0008', 'dev-seed', 'push', NOW() - INTERVAL '3 day', '2.21.0', 'payments integration suite', './integration', 25, 24, 1, 0, 0, 0, FALSE, FALSE, 25830, 'failed', 'prod'),

  ('e3330000-0000-0000-0000-000000000001', '33333333-3333-3333-3333-333333333333', 'main', 'webm0005', 'dev-seed', 'push', NOW() - INTERVAL '6 day', '2.21.0', 'webapp integration suite', './integration', 20, 19, 1, 0, 0, 0, FALSE, FALSE, 23310, 'failed', 'test'),
  ('e3330000-0000-0000-0000-000000000002', '33333333-3333-3333-3333-333333333333', 'main', 'webm0006', 'dev-seed', 'push', NOW() - INTERVAL '5 day', '2.21.0', 'webapp integration suite', './integration', 20, 20, 0, 0, 0, 0, FALSE, FALSE, 21940, 'passed', 'stage'),
  ('e3330000-0000-0000-0000-000000000003', '33333333-3333-3333-3333-333333333333', 'main', 'webm0007', 'dev-seed', 'push', NOW() - INTERVAL '4 day', '2.21.0', 'webapp integration suite', './integration', 20, 19, 1, 0, 0, 0, FALSE, FALSE, 24120, 'failed', 'prod'),
  ('e3330000-0000-0000-0000-000000000004', '33333333-3333-3333-3333-333333333333', 'main', 'webm0008', 'dev-seed', 'push', NOW() - INTERVAL '3 day', '2.21.0', 'webapp integration suite', './integration', 20, 20, 0, 0, 0, 0, FALSE, FALSE, 21490, 'passed', 'test'),

  ('e4440000-0000-0000-0000-000000000001', '44444444-4444-4444-4444-444444444444', 'main', 'autm0005', 'dev-seed', 'push', NOW() - INTERVAL '6 day', '2.21.0', 'auth-service integration suite', './integration', 22, 22, 0, 0, 0, 0, FALSE, FALSE, 22060, 'passed', 'stage'),
  ('e4440000-0000-0000-0000-000000000002', '44444444-4444-4444-4444-444444444444', 'main', 'autm0006', 'dev-seed', 'push', NOW() - INTERVAL '5 day', '2.21.0', 'auth-service integration suite', './integration', 22, 21, 1, 0, 0, 0, FALSE, FALSE, 23570, 'failed', 'prod'),
  ('e4440000-0000-0000-0000-000000000003', '44444444-4444-4444-4444-444444444444', 'main', 'autm0007', 'dev-seed', 'push', NOW() - INTERVAL '4 day', '2.21.0', 'auth-service integration suite', './integration', 22, 22, 0, 0, 0, 0, FALSE, FALSE, 21820, 'passed', 'test'),
  ('e4440000-0000-0000-0000-000000000004', '44444444-4444-4444-4444-444444444444', 'main', 'autm0008', 'dev-seed', 'push', NOW() - INTERVAL '3 day', '2.21.0', 'auth-service integration suite', './integration', 22, 21, 1, 0, 0, 0, FALSE, FALSE, 23210, 'failed', 'stage'),

  ('e5550000-0000-0000-0000-000000000001', '55555555-5555-5555-5555-555555555555', 'main', 'notm0005', 'dev-seed', 'push', NOW() - INTERVAL '6 day', '2.21.0', 'notifications integration suite', './integration', 16, 15, 1, 0, 0, 0, FALSE, FALSE, 18270, 'failed', 'prod'),
  ('e5550000-0000-0000-0000-000000000002', '55555555-5555-5555-5555-555555555555', 'main', 'notm0006', 'dev-seed', 'push', NOW() - INTERVAL '5 day', '2.21.0', 'notifications integration suite', './integration', 16, 16, 0, 0, 0, 0, FALSE, FALSE, 16990, 'passed', 'test'),
  ('e5550000-0000-0000-0000-000000000003', '55555555-5555-5555-5555-555555555555', 'main', 'notm0007', 'dev-seed', 'push', NOW() - INTERVAL '4 day', '2.21.0', 'notifications integration suite', './integration', 16, 15, 1, 0, 0, 0, FALSE, FALSE, 18810, 'failed', 'stage'),
  ('e5550000-0000-0000-0000-000000000004', '55555555-5555-5555-5555-555555555555', 'main', 'notm0008', 'dev-seed', 'push', NOW() - INTERVAL '3 day', '2.21.0', 'notifications integration suite', './integration', 16, 16, 0, 0, 0, 0, FALSE, FALSE, 16450, 'passed', 'prod'),

  ('e6660000-0000-0000-0000-000000000001', '66666666-6666-6666-6666-666666666666', 'main', 'catm0005', 'dev-seed', 'push', NOW() - INTERVAL '6 day', '2.21.0', 'catalog integration suite', './integration', 24, 23, 1, 0, 0, 0, FALSE, FALSE, 25820, 'failed', 'test'),
  ('e6660000-0000-0000-0000-000000000002', '66666666-6666-6666-6666-666666666666', 'main', 'catm0006', 'dev-seed', 'push', NOW() - INTERVAL '5 day', '2.21.0', 'catalog integration suite', './integration', 24, 24, 0, 0, 0, 0, FALSE, FALSE, 24410, 'passed', 'stage'),
  ('e6660000-0000-0000-0000-000000000003', '66666666-6666-6666-6666-666666666666', 'main', 'catm0007', 'dev-seed', 'push', NOW() - INTERVAL '4 day', '2.21.0', 'catalog integration suite', './integration', 24, 23, 1, 0, 0, 0, FALSE, FALSE, 26690, 'failed', 'prod'),
  ('e6660000-0000-0000-0000-000000000004', '66666666-6666-6666-6666-666666666666', 'main', 'catm0008', 'dev-seed', 'push', NOW() - INTERVAL '3 day', '2.21.0', 'catalog integration suite', './integration', 24, 24, 0, 0, 0, 0, FALSE, FALSE, 23950, 'passed', 'test'),

  ('e7770000-0000-0000-0000-000000000001', '77777777-7777-7777-7777-777777777777', 'main', 'ordm0005', 'dev-seed', 'push', NOW() - INTERVAL '6 day', '2.21.0', 'orders integration suite', './integration', 14, 14, 0, 0, 0, 0, FALSE, FALSE, 16830, 'passed', 'stage'),
  ('e7770000-0000-0000-0000-000000000002', '77777777-7777-7777-7777-777777777777', 'main', 'ordm0006', 'dev-seed', 'push', NOW() - INTERVAL '5 day', '2.21.0', 'orders integration suite', './integration', 14, 13, 1, 0, 0, 0, FALSE, FALSE, 18200, 'failed', 'prod'),
  ('e7770000-0000-0000-0000-000000000003', '77777777-7777-7777-7777-777777777777', 'main', 'ordm0007', 'dev-seed', 'push', NOW() - INTERVAL '4 day', '2.21.0', 'orders integration suite', './integration', 14, 14, 0, 0, 0, 0, FALSE, FALSE, 16690, 'passed', 'test'),
  ('e7770000-0000-0000-0000-000000000004', '77777777-7777-7777-7777-777777777777', 'main', 'ordm0008', 'dev-seed', 'push', NOW() - INTERVAL '3 day', '2.21.0', 'orders integration suite', './integration', 14, 13, 1, 0, 0, 0, FALSE, FALSE, 17940, 'failed', 'stage'),

  ('e8880000-0000-0000-0000-000000000001', '88888888-8888-8888-8888-888888888888', 'main', 'shpm0005', 'dev-seed', 'push', NOW() - INTERVAL '6 day', '2.21.0', 'shipping integration suite', './integration', 15, 14, 1, 0, 0, 0, FALSE, FALSE, 18640, 'failed', 'prod'),
  ('e8880000-0000-0000-0000-000000000002', '88888888-8888-8888-8888-888888888888', 'main', 'shpm0006', 'dev-seed', 'push', NOW() - INTERVAL '5 day', '2.21.0', 'shipping integration suite', './integration', 15, 15, 0, 0, 0, 0, FALSE, FALSE, 17130, 'passed', 'test'),
  ('e8880000-0000-0000-0000-000000000003', '88888888-8888-8888-8888-888888888888', 'main', 'shpm0007', 'dev-seed', 'push', NOW() - INTERVAL '4 day', '2.21.0', 'shipping integration suite', './integration', 15, 14, 1, 0, 0, 0, FALSE, FALSE, 19020, 'failed', 'stage'),
  ('e8880000-0000-0000-0000-000000000004', '88888888-8888-8888-8888-888888888888', 'main', 'shpm0008', 'dev-seed', 'push', NOW() - INTERVAL '3 day', '2.21.0', 'shipping integration suite', './integration', 15, 15, 0, 0, 0, 0, FALSE, FALSE, 16790, 'passed', 'prod'),

  ('e9990000-0000-0000-0000-000000000001', '99999999-9999-9999-9999-999999999999', 'main', 'anam0005', 'dev-seed', 'push', NOW() - INTERVAL '6 day', '2.21.0', 'analytics integration suite', './integration', 17, 16, 1, 0, 0, 0, FALSE, FALSE, 20540, 'failed', 'test'),
  ('e9990000-0000-0000-0000-000000000002', '99999999-9999-9999-9999-999999999999', 'main', 'anam0006', 'dev-seed', 'push', NOW() - INTERVAL '5 day', '2.21.0', 'analytics integration suite', './integration', 17, 17, 0, 0, 0, 0, FALSE, FALSE, 19120, 'passed', 'stage'),
  ('e9990000-0000-0000-0000-000000000003', '99999999-9999-9999-9999-999999999999', 'main', 'anam0007', 'dev-seed', 'push', NOW() - INTERVAL '4 day', '2.21.0', 'analytics integration suite', './integration', 17, 16, 1, 0, 0, 0, FALSE, FALSE, 21410, 'failed', 'prod'),
  ('e9990000-0000-0000-0000-000000000004', '99999999-9999-9999-9999-999999999999', 'main', 'anam0008', 'dev-seed', 'push', NOW() - INTERVAL '3 day', '2.21.0', 'analytics integration suite', './integration', 17, 17, 0, 0, 0, 0, FALSE, FALSE, 18780, 'passed', 'test'),

  ('eaaa0000-0000-0000-0000-000000000001', 'aaaaaaaa-1111-1111-1111-111111111111', 'main', 'seam0005', 'dev-seed', 'push', NOW() - INTERVAL '6 day', '2.21.0', 'search integration suite', './integration', 21, 21, 0, 0, 0, 0, FALSE, FALSE, 23150, 'passed', 'stage'),
  ('eaaa0000-0000-0000-0000-000000000002', 'aaaaaaaa-1111-1111-1111-111111111111', 'main', 'seam0006', 'dev-seed', 'push', NOW() - INTERVAL '5 day', '2.21.0', 'search integration suite', './integration', 21, 20, 1, 0, 0, 0, FALSE, FALSE, 24670, 'failed', 'prod'),
  ('eaaa0000-0000-0000-0000-000000000003', 'aaaaaaaa-1111-1111-1111-111111111111', 'main', 'seam0007', 'dev-seed', 'push', NOW() - INTERVAL '4 day', '2.21.0', 'search integration suite', './integration', 21, 21, 0, 0, 0, 0, FALSE, FALSE, 22890, 'passed', 'test'),
  ('eaaa0000-0000-0000-0000-000000000004', 'aaaaaaaa-1111-1111-1111-111111111111', 'main', 'seam0008', 'dev-seed', 'push', NOW() - INTERVAL '3 day', '2.21.0', 'search integration suite', './integration', 21, 20, 1, 0, 0, 0, FALSE, FALSE, 24220, 'failed', 'stage')
ON CONFLICT (id) DO NOTHING;

INSERT INTO integration_spec_results (
  id, integration_run_id, spec_path, leaf_node_text, state, duration_ms,
  failure_message, failure_location_file, failure_location_line
)
VALUES
  ('d4000000-0000-0000-0000-000000000001', 'c4000000-0000-0000-0000-000000000001', 'OIDC > refresh token rotation', 'refresh token rotation', 'passed', 131, NULL, NULL, NULL),
  ('d4000000-0000-0000-0000-000000000002', 'c4000000-0000-0000-0000-000000000001', 'OIDC > maps custom claim groups', 'maps custom claim groups', 'failed', 92, 'expected groups claim to include team-admin', 'integration/oidc_claims_test.go', 118),
  ('d4000000-0000-0000-0000-000000000003', 'c4000000-0000-0000-0000-000000000002', 'RBAC > validates service token', 'validates service token', 'passed', 84, NULL, NULL, NULL),
  ('d4000000-0000-0000-0000-000000000004', 'c4000000-0000-0000-0000-000000000002', 'RBAC > denies unknown scope', 'denies unknown scope', 'skipped', 0, NULL, NULL, NULL),

  ('d5000000-0000-0000-0000-000000000001', 'c5000000-0000-0000-0000-000000000001', 'Email > sends templated digest', 'sends templated digest', 'passed', 122, NULL, NULL, NULL),
  ('d5000000-0000-0000-0000-000000000002', 'c5000000-0000-0000-0000-000000000001', 'Retry > exponential backoff with jitter', 'exponential backoff with jitter', 'flaky', 211, 'transient broker timeout recovered on retry', 'integration/retry_test.go', 67),
  ('d5000000-0000-0000-0000-000000000003', 'c5000000-0000-0000-0000-000000000001', 'Webhook > signs outbound payload', 'signs outbound payload', 'failed', 77, 'expected HMAC header to be present', 'integration/webhook_test.go', 91),
  ('d5000000-0000-0000-0000-000000000004', 'c5000000-0000-0000-0000-000000000002', 'Batch > drains queue on shutdown', 'drains queue on shutdown', 'passed', 98, NULL, NULL, NULL),

  ('d6000000-0000-0000-0000-000000000001', 'c6000000-0000-0000-0000-000000000001', 'Ingest > merges duplicate SKU updates', 'merges duplicate SKU updates', 'passed', 126, NULL, NULL, NULL),
  ('d6000000-0000-0000-0000-000000000002', 'c6000000-0000-0000-0000-000000000001', 'Search > reindex on schema migration', 'reindex on schema migration', 'failed', 101, 'expected zero stale docs after migration', 'integration/reindex_test.go', 144),
  ('d6000000-0000-0000-0000-000000000003', 'c6000000-0000-0000-0000-000000000001', 'Stock > backfill historical events', 'backfill historical events', 'pending', 0, NULL, NULL, NULL),
  ('d6000000-0000-0000-0000-000000000004', 'c6000000-0000-0000-0000-000000000002', 'Stock > syncs external inventory snapshot', 'syncs external inventory snapshot', 'passed', 115, NULL, NULL, NULL),

  ('d8000000-0000-0000-0000-000000000001', 'c8000000-0000-0000-0000-000000000001', 'Labels > caches generated labels', 'caches generated labels', 'passed', 110, NULL, NULL, NULL),
  ('d8000000-0000-0000-0000-000000000002', 'c8000000-0000-0000-0000-000000000001', 'Rates > fallback carrier lookup', 'fallback carrier lookup', 'failed', 86, 'expected fallback carrier response within SLA', 'integration/rates_test.go', 79),
  ('d8000000-0000-0000-0000-000000000003', 'c8000000-0000-0000-0000-000000000002', 'Labels > cache invalidation on template change', 'cache invalidation on template change', 'passed', 97, NULL, NULL, NULL),

  ('d9000000-0000-0000-0000-000000000001', 'c9000000-0000-0000-0000-000000000001', 'ETL > backfills historical partitions', 'backfills historical partitions', 'passed', 132, NULL, NULL, NULL),
  ('d9000000-0000-0000-0000-000000000002', 'c9000000-0000-0000-0000-000000000001', 'Dashboards > enforces tenant filters', 'enforces tenant filters', 'failed', 95, 'expected tenant filter in generated SQL', 'integration/dashboards_test.go', 103),
  ('d9000000-0000-0000-0000-000000000003', 'c9000000-0000-0000-0000-000000000001', 'Dashboards > recovers stale cache snapshots', 'recovers stale cache snapshots', 'pending', 0, NULL, NULL, NULL),

  ('da000000-0000-0000-0000-000000000001', 'ca000000-0000-0000-0000-000000000001', 'Ranking > recalculates BM25 boosts', 'recalculates BM25 boosts', 'passed', 123, NULL, NULL, NULL),
  ('da000000-0000-0000-0000-000000000002', 'ca000000-0000-0000-0000-000000000001', 'Index > handles alias swap with no downtime', 'handles alias swap with no downtime', 'passed', 119, NULL, NULL, NULL)
ON CONFLICT (id) DO NOTHING;
