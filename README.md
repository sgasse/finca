# Financial Calculations (FinCa)

Simulation of passive investment strategies with publicly available data from [AlphaVantage](https://www.alphavantage.co/), visualized with [Apache ECharts](https://echarts.apache.org/en/index.html). This project was written in golang for self-educational purposes.

### Background
I am a fan of passive investment: Regularily growing a diversified portfolio of [exchange traded funds (ETFs)](https://en.wikipedia.org/wiki/Exchange-traded_fund). Passive investment accepts that non-professional investors will statistically not be able to outperform the general stock market over long periods of time. This is based on the [efficient market hypothesis](https://en.wikipedia.org/wiki/Efficient-market_hypothesis). Thus I normally only put money into the market, buying new stocks, never selling as an active investor would do.
Nevertheless, the volatility through the recent crisises made me wonder what the best strategy for investing would be.

### Considerations
There were two major questions which I wanted to answer:
 - Could it be advantageous to keep your money in cash until a dramatic down-turn of the market happens? Could this outperform a strategy which invests money periodically regardless of the current state of the market?
 - Given certain costs to investment, what difference does it make to invest every month compared to e.g. twice a year?

### My personal situation
As you can imagine, I do not want to share all details of my financial situation here. But I can say that I currently save a certain fixed amount of my income every month. At (so far) irregular intervals, I rebalanced my existing ETF stock portfolio by investing what I saved up until then.
However lately, I sometimes wondered 'Would it be smarter to wait a little before investing?'. There is the saying 'Time in the market beats timing the market' and I imagine that there is some wisdom in it. Still, where is the lower limit to it? Should I invest every month? And I also wanted to give myself some peace of mind with hard numbers.

### The problem with free financial data
In the beginning, I just planned to write a simulation that would rebalance my actual ETF portfolio given different strategies for longer historic periods. Though I encountered some problems that made me change my approach:
 - Most of the ETFs which I have in my portfolio do not exist for longer than 6-10 years. This seems to short for me to really base an analysis on.
 - Free stock data is not as easy to get as I imagined. In the golden years of the Yahoo! Finance API, it was substantially easier to work with historic stock market data.
After some searching, I decided to make two simplifications which I think will not impact the validity of the general findings:
 - Instead of simulating and rebalancing a complete portfolio, I instead use a single market index ETF as stand-in for a realistic development of the stock market over a longer period of time.
 - The costs per investment which I take into account are the costs which I would pay for rebalancing my real portfolio.
 - The historical end-of-day data which I take into account comes from [AlphaVantage](https://www.alphavantage.co/) and spans a period of 20 years. Ideally, it would be longer, but I did not find free daily data reaching back longer. Ping me if you have free daily EOD data to share :)
