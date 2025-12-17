import React, { createContext, useContext, useState } from 'react';

const ContentsPanelContext = createContext({
    contentsComponent: null,
    setContentsComponent: () => {},
    contentsData: null,
    setContentsData: () => {}
});

export const ContentsPanelProvider = ({ children }) => {
    const [contentsComponent, setContentsComponent] = useState(null);
    const [contentsData, setContentsData] = useState(null);

    return (
        <ContentsPanelContext.Provider 
            value={{
                contentsComponent,
                setContentsComponent,
                contentsData,
                setContentsData
            }}
        >
            {children}
        </ContentsPanelContext.Provider>
    );
};

export const useContentsPanel = () => {
    const context = useContext(ContentsPanelContext);
    if (!context) {
        throw new Error('useContentsPanel must be used within ContentsPanelProvider');
    }
    return context;
};
