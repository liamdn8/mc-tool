import React, { useState, useRef, useEffect } from 'react';
import { ChevronDown, ChevronUp } from 'lucide-react';

const PanelGroup = ({ children, direction = 'vertical' }) => {
    const [panelSizes, setPanelSizes] = useState([75, 25]); // Initial split: 75% top, 25% bottom
    const [isResizing, setIsResizing] = useState(false);
    const containerRef = useRef(null);
    const startPosRef = useRef(0);
    const startSizesRef = useRef([]);

    const panels = React.Children.toArray(children).filter(
        child => child.type === Panel
    );

    const handleMouseDown = (e, index) => {
        e.preventDefault();
        setIsResizing(true);
        startPosRef.current = direction === 'vertical' ? e.clientY : e.clientX;
        startSizesRef.current = [...panelSizes];
        
        document.body.style.cursor = direction === 'vertical' ? 'row-resize' : 'col-resize';
        document.body.style.userSelect = 'none';
    };

    useEffect(() => {
        if (!isResizing) return;

        const handleMouseMove = (e) => {
            if (!containerRef.current) return;

            const container = containerRef.current;
            const containerSize = direction === 'vertical' 
                ? container.offsetHeight 
                : container.offsetWidth;
            
            const currentPos = direction === 'vertical' ? e.clientY : e.clientX;
            const delta = currentPos - startPosRef.current;
            const deltaPercent = (delta / containerSize) * 100;

            const newSizes = [...startSizesRef.current];
            newSizes[0] = Math.max(10, Math.min(90, startSizesRef.current[0] + deltaPercent));
            newSizes[1] = 100 - newSizes[0];

            setPanelSizes(newSizes);
        };

        const handleMouseUp = () => {
            setIsResizing(false);
            document.body.style.cursor = '';
            document.body.style.userSelect = '';
        };

        document.addEventListener('mousemove', handleMouseMove);
        document.addEventListener('mouseup', handleMouseUp);

        return () => {
            document.removeEventListener('mousemove', handleMouseMove);
            document.removeEventListener('mouseup', handleMouseUp);
        };
    }, [isResizing, direction]);

    return (
        <div 
            ref={containerRef}
            className={`panel-group panel-group-${direction}`}
            data-panel-group=""
            data-panel-group-direction={direction}
        >
            {panels.map((panel, index) => (
                <React.Fragment key={index}>
                    {React.cloneElement(panel, {
                        size: panelSizes[index],
                        direction
                    })}
                    {index < panels.length - 1 && (
                        <div
                            className={`panel-resize-handle panel-resize-handle-${direction}`}
                            data-resize-handle=""
                            onMouseDown={(e) => handleMouseDown(e, index)}
                        >
                            <div className="resize-handle-bar" />
                        </div>
                    )}
                </React.Fragment>
            ))}
        </div>
    );
};

const Panel = ({ 
    id, 
    title, 
    children, 
    collapsible = true, 
    size = 50, 
    direction,
    onToggleCollapse,
    collapsed = false
}) => {
    const [isCollapsed, setIsCollapsed] = useState(collapsed);

    const toggleCollapse = () => {
        const newState = !isCollapsed;
        setIsCollapsed(newState);
        if (onToggleCollapse) {
            onToggleCollapse(newState);
        }
    };

    return (
        <div
            id={id}
            className={`panel ${isCollapsed ? 'panel-collapsed' : ''}`}
            data-panel=""
            data-panel-collapsible={collapsible}
            data-panel-id={id}
            data-panel-size={size}
            style={{
                flex: isCollapsed ? '0 0 auto' : `${size} 1 0px`,
                overflow: 'hidden'
            }}
        >
            {title && (
                <div className="panel-header">
                    {collapsible ? (
                        <button 
                            className="panel-expand-collapse-btn"
                            onClick={toggleCollapse}
                            aria-label={isCollapsed ? `Show ${title}` : `Hide ${title}`}
                        >
                            {isCollapsed ? (
                                <ChevronUp size={16} />
                            ) : (
                                <ChevronDown size={16} />
                            )}
                            <h4 className="panel-title">{title}</h4>
                        </button>
                    ) : (
                        <h4 className="panel-title">{title}</h4>
                    )}
                </div>
            )}
            <div className={`panel-body ${isCollapsed ? 'panel-body-hidden' : ''}`}>
                {children}
            </div>
        </div>
    );
};

PanelGroup.Panel = Panel;

export default PanelGroup;
export { Panel };
