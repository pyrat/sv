import winston, { Logger as WinstonLogger, format, transports } from 'winston';

class Logger {
    private logger: WinstonLogger;

    constructor(){
        this.logger = winston.createLogger({
            level: 'silly', 
            format: format.combine(
                format.timestamp({
                    format: 'DD-MM-YYYY HH:mm:ss'
                }),
                format.errors({ stack: true }),
                format.splat(),
                format.json()
            ),
            transports: [
                new transports.File({ filename: 'logs/daWinchi-info.log', level: 'info'}),
                new transports.File({ filename: 'logs/daWinchi-error.log', level: 'error' }),
                new transports.File({ filename: 'logs/daWinchi-debug.log', level: 'debug'}),
                new transports.File({ filename: 'logs/daWinchi-silly.log'})
            ]
        });

        if (process.env.NODE_ENV !== 'production') {
            this.logger.add(new transports.Console({
              format: format.combine(
                format.colorize(),
                format.simple()
              )
            }));
        }
    }

    /**
     * @param {string} level - Level of priority. error, warn, info, http, verbose, debug, silly
     * @param {string} message 
     */
    log(level: string, message: string, ...args: any[]){
        this.logger.log(level, message, ...args);
    }
}

export default new Logger();
