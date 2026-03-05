import { parseAnyRPCResponse, RPCChannelStatus, RPCMethod, RPCResponse } from '@layer-3/nitrolite';
import { WebSocket } from 'ws';

interface MessageListener {
    id: string;
    callback: (data: string) => void;
}

export class TestWebSocket {
    private socket: WebSocket | null = null;
    private messageListeners: MessageListener[] = [];
    private closed = false;

    constructor(private url: string, private debugMode = false) {}

    connect(): Promise<void> {
        return new Promise((resolve, reject) => {
            if (this.socket) {
                return resolve();
            }

            this.socket = new WebSocket(this.url);
            this.closed = false;

            this.socket.on('open', () => {
                if (this.debugMode) {
                    console.log('WebSocket connection established');
                }

                resolve();
            });

            this.socket.on('message', (event) => {
                if (this.closed) {
                    return;
                }

                if (this.debugMode) {
                    console.log('Message received:', event.toString());
                }

                for (const listener of this.messageListeners) {
                    try {
                        listener.callback(event);
                    } catch (error) {
                        console.error('Error in message listener:', error);
                    }
                }
            });

            this.socket.on('error', (error) => {
                if (this.closed) {
                    return;
                }

                if (this.debugMode && !this.closed) {
                    console.error('WebSocket error:', error);
                }

                reject(error);
            });

            this.socket.on('close', () => {
                this.socket = null;
            });
        });
    }

    waitForMessage(predicate: (data: string, reqId?: number) => boolean, reqId?: number, timeout = 1000): Promise<string> {
        return new Promise((resolve, reject) => {
            const listenerId = Math.random().toString(36).substring(2);

            const timeoutId = setTimeout(() => {
                this.removeMessageListener(listenerId);
                reject(new Error(`Timeout waiting for message after ${timeout}ms. Request ID: ${reqId || 'N/A'}`));
            }, timeout);

            const messageHandler = (data: string) => {
                try {
                    if (this.closed) {
                        clearTimeout(timeoutId);
                        this.removeMessageListener(listenerId);
                        return;
                    }

                    if (predicate(data, reqId)) {
                        clearTimeout(timeoutId);
                        this.removeMessageListener(listenerId);
                        resolve(data);
                    }
                } catch (error) {
                    clearTimeout(timeoutId);
                    this.removeMessageListener(listenerId);
                    reject(new Error(`Error in predicate function: ${error}`));
                }
            };

            this.addMessageListener(listenerId, messageHandler);
        });
    }

    private addMessageListener(id: string, callback: (data: string) => void): void {
        this.messageListeners.push({ id, callback });
    }

    private removeMessageListener(id: string): void {
        this.messageListeners = this.messageListeners.filter((listener) => listener.id !== id);
    }

    send(message: string): void {
        if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
            throw new Error('WebSocket is not open. Cannot send message.');
        }

        if (this.debugMode) {
            console.log('Sending message:', message);
        }

        this.socket.send(message);
    }

    sendAndWaitForResponse(message: string, predicate: (data: string) => boolean, timeout = 1000): Promise<any> {
        const parsedMessage = JSON.parse(message);
        const reqId = parsedMessage.req[0];

        const messagePromise = this.waitForMessage(predicate, reqId, timeout);
        this.send(message);
        return messagePromise;
    }

    close(): void {
        if (this.socket) {
            this.socket.close();
            this.socket = null;
            this.messageListeners = [];
            this.closed = true;
        }
    }
}

const genericPredicate = (data: string, condition: (r: RPCResponse) => boolean, reqId?: number) => {
    try {
        const parsedData = parseAnyRPCResponse(data);
        if (condition(parsedData)) {
            return true;
        }

        if (reqId !== undefined && parsedData.requestId === reqId && parsedData.method === RPCMethod.Error) {
            const errorMsg = parsedData.params.error;
            throw new Error(`RPC Error: ${errorMsg}`);
        }
    } catch (error) {
        if (error.message.includes('Unsupported RPC method: assets')) {
            return false; // TODO: Ignore unsupported method errors
        }

        throw new Error(`Error parsing data: ${error.message}. Data: ${data}`);
    }

    return false;
};

export const getPongPredicate = () => {
    return (data: string, reqId?: number): boolean => {
        return genericPredicate(data, (r) => r.method === RPCMethod.Pong, reqId);
    };
};

export const getAuthChallengePredicate = () => {
    return (data: string, reqId?: number): boolean => {
        return genericPredicate(data, (r) => r.method === RPCMethod.AuthChallenge, reqId);
    };
};

export const getAuthVerifyPredicate = () => {
    return (data: string, reqId?: number): boolean => {
        return genericPredicate(data, (r) => r.method === RPCMethod.AuthVerify, reqId);
    };
};

export const getChannelUpdatePredicateWithStatus = (status: RPCChannelStatus) => {
    return (data: string, reqId?: number): boolean => {
        return genericPredicate(
            data,
            (r) => {
                return r.method === RPCMethod.ChannelUpdate && r.params.status === status;
            },
            reqId
        );
    };
};

export const getCreateChannelPredicate = () => {
    return (data: string, reqId?: number): boolean => {
        return genericPredicate(data, (r) => r.method === RPCMethod.CreateChannel, reqId);
    }
};

export const getCloseChannelPredicate = () => {
    return (data: string, reqId?: number): boolean => {
        return genericPredicate(data, (r) => r.method === RPCMethod.CloseChannel, reqId);
    };
};

export const getResizeChannelPredicate = () => {
    return (data: string, reqId?: number): boolean => {
        return genericPredicate(data, (r) => r.method === RPCMethod.ResizeChannel, reqId);
    };
};

export const getErrorPredicate = () => {
    return (data: string, _?: number): boolean => {
        // No need for reqId here, as we are checking for any error response
        return genericPredicate(data, (r) => r.method === RPCMethod.Error);
    };
};

export const getGetLedgerBalancesPredicate = () => {
    return (data: string, reqId?: number): boolean => {
        return genericPredicate(data, (r) => r.method === RPCMethod.GetLedgerBalances, reqId);
    };
};

export const getGetLedgerEntriesPredicate = () => {
    return (data: string, reqId?: number): boolean => {
        return genericPredicate(data, (r) => r.method === RPCMethod.GetLedgerEntries, reqId);
    };
};

export const getGetLedgerTransactionsPredicate = () => {
    return (data: string, reqId?: number): boolean => {
        return genericPredicate(data, (r) => r.method === RPCMethod.GetLedgerTransactions, reqId);
    };
};

export const getGetSessionKeysPredicate = () => {
    return (data: string, reqId?: number): boolean => {
        return genericPredicate(data, (r) => r.method === RPCMethod.GetSessionKeys, reqId);
    };
}

export const getRevokeSessionKeyPredicate = () => {
    return (data: string, reqId?: number): boolean => {
        return genericPredicate(data, (r) => r.method === RPCMethod.RevokeSessionKey, reqId);
    };
};

export const getTransferPredicate = () => {
    return (data: string, reqId?: number): boolean => {
        return genericPredicate(data, (r) => r.method === RPCMethod.Transfer, reqId);
    };
};

export const getCreateAppSessionPredicate = () => {
    return (data: string, reqId?: number): boolean => {
        return genericPredicate(data, (r) => r.method === RPCMethod.CreateAppSession, reqId);
    };
};

export const getSubmitAppStatePredicate = () => {
    return (data: string, reqId?: number): boolean => {
        return genericPredicate(data, (r) => r.method === RPCMethod.SubmitAppState, reqId);
    };
};

export const getGetAppSessionsPredicate = () => {
    return (data: string, reqId?: number): boolean => {
        return genericPredicate(data, (r) => r.method === RPCMethod.GetAppSessions, reqId);
    };
};

export const getCloseAppSessionPredicate = () => {
    return (data: string, reqId?: number): boolean => {
        return genericPredicate(data, (r) => r.method === RPCMethod.CloseAppSession, reqId);
    };
};

export const getCleanupSessionKeyCachePredicate = () => {
    return (data: string, reqId?: number): boolean => {
        return genericPredicate(data, (r) => r.method === RPCMethod.CleanupSessionKeyCache, reqId);
    };
};
