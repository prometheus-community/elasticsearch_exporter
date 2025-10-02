local g = import './g.libsonnet';
local prometheusQuery = g.query.prometheus;

local variables = import './variables.libsonnet';

(import './queries/general.libsonnet') +
(import './queries/shard.libsonnet') +
(import './queries/document.libsonnet') +
(import './queries/memory.libsonnet') +
(import './queries/threads.libsonnet') +
(import './queries/network.libsonnet')
