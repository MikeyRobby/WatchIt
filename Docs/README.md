
Software for the watch lovers to track their collection and for the hobbyist and independent shops to track their work.
# Overview
WatchIt was created out of two loves of mine; Watches and Programming. WatchIt is split into 2 pieces, the Monitor is designed to track the prices of watches across multiple marketplaces designed with buying/selling watches in mind. The toolbox is the other half, designed with repair/servicing in mind. Features of the toolbox include tracking service orders, invoices, work done, and report generation showing all of the work done. Take a look at the [[WatchIt/Docs/Design Document|Design Document]] for a breakdown of why I built WatchIt and the [[WatchIt/Docs/Technical Document|Technical Document]] for a breakdown of how I built WatchIt.

# Monitor

Monitor is designed primarily to track the prices and price movement of watches in your collection. Data is pulled from eBay and monitored over time to show spot prices and previous prices with trends being shown on a graph. Watchlist can be created for potential buys and collections can be created to track the value of your pieces.  You can view your watchlist and collections on a dashboard showing the average price of each piece across all of the sources.

# Toolbox
Toolbox has features designed for hobbyist and and small shops to use for repair/servicing and custom builds for people. Key features are service orders, invoicing, tracking work done, and generating reports for record keeping and sending out to clients. Each client will have a profile that will show all of the work done for that client and help keep track of return clients.

## Work Orders
You can create a work order for clients with detailed descriptions of what is expected. During the process of servicing the watches photos can be attached to the order along with details per photo showcasing the steps completed. When reports are generated these photos and descriptions will be put together to give a final detailed showcase of everything you've done for the client that can be shared with the client and saved to the client profile for record keeping.

## Reporting
After work has been completed you can generate a report showcasing everything that was done for the client, work orders, and invoices. The reports help with record keeping and any disputes that may come up during or after the job is complete. Reports can be generated into PDFs for easy viewing on any device.

# Technology
WatchIt is written primarily using Go as the backend with Wails as the front-end and sqlLite as a Database. Check out the [[WatchIt/Docs/Technical Document|Technical Document]]  for an in-depth breakdown of the technologies used.

# Features Implemented
	[ ] Monitor: Price Tracking
	[ ] Monitor: Dashboard
	[ ] Monitor: Watchlist
	[ ] Monitor: Collections
	[ ] Toolbox: Work Order
	[ ] Toolbox: Invoices
	[ ] Toolbox: Client Profile
	[ ] Toolbox: Report Generation