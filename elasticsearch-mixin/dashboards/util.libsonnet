local g = import 'g.libsonnet';
local panelUtil = g.util.panel;

{
  local gridWidth = 24,

  // makeGrid returns an array of panels organized into a grid layout.
  // This is a modified version of the grafonnet makeGrid function to
  // calculate the width of each panel based on the number of panels.
  makeGrid(panels, panelHeight=4, startY=0):
    local sanitizePanels(ps) =
      // Figure out the number of panels and the width of each panel
      local numPanels = std.length(ps);
      local panelWidth = std.floor(gridWidth / numPanels);

      // Sanitize the panels, this ensures tht the panels have the valid gridPos
      std.map(
        function(p)
          local sanePanel = panelUtil.sanitizePanel(p, defaultHeight=panelHeight);
          (
            if p.type == 'row'
            then sanePanel {
              panels: sanitizePanels(sanePanel.panels),
            }
            else sanePanel {
              gridPos+: {
                w: panelWidth,
              },
            }
          ),
        ps
      );

    local sanitizedPanels = sanitizePanels(panels);

    local grouped = panelUtil.groupPanelsInRows(sanitizedPanels);

    local panelsBeforeRows = panelUtil.getPanelsBeforeNextRow(grouped);
    local rowPanels =
      std.filter(
        function(p) p.type == 'row',
        grouped
      );


    local CalculateXforPanel(index, panel) =
      local panelsPerRow = std.floor(gridWidth / panel.gridPos.w);
      local col = std.mod(index, panelsPerRow);
      panel { gridPos+: { x: panel.gridPos.w * col } };


    local panelsBeforeRowsWithX = std.mapWithIndex(CalculateXforPanel, panelsBeforeRows);

    local rowPanelsWithX =
      std.map(
        function(row)
          row { panels: std.mapWithIndex(CalculateXforPanel, row.panels) },
        rowPanels
      );

    local uncollapsed = panelUtil.resolveCollapsedFlagOnRows(panelsBeforeRowsWithX + rowPanelsWithX);

    local normalized = panelUtil.normalizeY(uncollapsed);

    std.map(function(p) p { gridPos+: { y+: startY } }, normalized),
}
