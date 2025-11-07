import React, { useEffect, useRef, useState } from 'react';
import { Terminal as XTerm } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';
import { useI18n } from '../utils/i18n';

const statusClass = (state) => {
    switch (state) {
        case 'connected':
            return 'badge-success';
        case 'error':
            return 'badge-danger';
        case 'disconnected':
            return 'badge-warning';
        default:
            return 'badge-warning';
    }
};

const TerminalPage = () => {
    const { t } = useI18n();
    const connectionLabels = {
        connecting: t('terminal_connecting', 'Connecting…'),
        connected: t('terminal_connected', 'Connected'),
        error: t('terminal_error', 'Error'),
        disconnected: t('terminal_disconnected', 'Disconnected')
    };
    const containerRef = useRef(null);
    const terminalRef = useRef(null);
    const fitAddonRef = useRef(null);
    const socketRef = useRef(null);

    const [connectionState, setConnectionState] = useState('connecting');
    const [errorMessage, setErrorMessage] = useState('');

    useEffect(() => {
        const term = new XTerm({
            cursorBlink: true,
            fontFamily: 'Menlo, Monaco, "Courier New", monospace',
            fontSize: 14,
            allowProposedApi: true,
            theme: {
                background: '#111111',
                foreground: '#EDEDED',
                cursor: '#EDEDED'
            }
        });
        const fitAddon = new FitAddon();

        terminalRef.current = term;
        fitAddonRef.current = fitAddon;

        term.loadAddon(fitAddon);
        term.open(containerRef.current);
        fitAddon.fit();
        term.focus();

        const sendResize = () => {
            if (!socketRef.current || socketRef.current.readyState !== WebSocket.OPEN) {
                return;
            }

            socketRef.current.send(JSON.stringify({
                type: 'resize',
                cols: term.cols,
                rows: term.rows
            }));
        };

        const handleResize = () => {
            if (!fitAddonRef.current) {
                return;
            }
            fitAddonRef.current.fit();
            sendResize();
        };

        const resizeObserver = typeof ResizeObserver !== 'undefined' ? new ResizeObserver(handleResize) : null;
        if (resizeObserver && containerRef.current) {
            resizeObserver.observe(containerRef.current);
        }

        const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
        const socket = new WebSocket(`${protocol}://${window.location.host}/api/terminal/ws`);
        socketRef.current = socket;

        term.onData((data) => {
            if (socket.readyState === WebSocket.OPEN) {
                socket.send(JSON.stringify({ type: 'input', data }));
            }
        });

        socket.onopen = () => {
            setConnectionState('connected');
            setErrorMessage('');
            fitAddon.fit();
            sendResize();
        };

        socket.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                switch (message.type) {
                    case 'output':
                        if (message.data) {
                            term.write(message.data);
                        }
                        break;
                    case 'error':
                        if (message.message) {
                            setErrorMessage(message.message);
                            term.writeln(`\r\n${message.message}`);
                        }
                        break;
                    case 'exit':
                        setConnectionState('disconnected');
                        if (typeof message.code === 'number') {
                            term.writeln(`\r\n${t('terminal_exit_with_code', 'Process exited with code {code}', { code: message.code })}`);
                        } else {
                            term.writeln(`\r\n${t('terminal_exit', 'Process exited')}`);
                        }
                        break;
                    case 'ready':
                        term.writeln(t('terminal_ready', 'Connected to host shell.'));
                        break;
                    default:
                        break;
                }
            } catch (err) {
                console.error('Failed to parse terminal message', err);
            }
        };

        socket.onerror = () => {
            setConnectionState('error');
            setErrorMessage(t('terminal_connection_error', 'Connection error'));
        };

        socket.onclose = () => {
            setConnectionState('disconnected');
        };

        window.addEventListener('resize', handleResize);

        return () => {
            window.removeEventListener('resize', handleResize);
            if (resizeObserver) {
                resizeObserver.disconnect();
            }

            if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
                socketRef.current.close();
            }

            term.dispose();
        };
    }, []);

    return (
        <div className="terminal-page">
            <div className="card">
                <div className="card-header">
                    <div>
                        <h2 className="card-title">{t('terminal_title', 'Live Terminal')}</h2>
                        <p className="card-subtitle">{t('terminal_subtitle', 'Run trusted commands on the mc-tool host')}</p>
                    </div>
                    <div className={`badge ${statusClass(connectionState)}`}>
                        {connectionLabels[connectionState] || connectionState}
                    </div>
                </div>
                <div className="terminal-wrapper">
                    <div ref={containerRef} className="terminal-container" />
                </div>
                <p className="terminal-disclaimer">
                    {t('terminal_disclaimer', 'Commands execute with the mc-tool service permissions. Use responsibly.')}
                    {errorMessage && (
                        <span className="terminal-error">{` ${errorMessage}`}</span>
                    )}
                </p>
            </div>
        </div>
    );
};

export default TerminalPage;
