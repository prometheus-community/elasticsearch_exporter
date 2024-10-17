local g = import 'g.libsonnet';

{
  stat: {
    local stat = g.panel.stat,

    base(title, targets):
      stat.new(title)
      + stat.queryOptions.withTargets(targets),

    nodes: self.base,
  },

  timeSeries: {
    local timeSeries = g.panel.timeSeries,

    base(title, targets):
      timeSeries.new(title)
      + timeSeries.queryOptions.withTargets(targets),

    ratio(title, targets):
      self.base(title, targets)
      + timeSeries.standardOptions.withUnit('percentunit'),

    ratioMax1(title, targets):
      self.ratio(title, targets)
      + timeSeries.standardOptions.withMax(1)
      + timeSeries.standardOptions.withMin(0),

    bytes(title, targets):
      self.base(title, targets)
      + timeSeries.standardOptions.withUnit('bytes'),

    seconds(title, targets):
      self.base(title, targets)
      + timeSeries.standardOptions.withUnit('s'),
  },
}
