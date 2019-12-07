import React from 'react';

export class Error extends React.Component {
    render() {
        return (
            <div id="error-content">
                {appData.data.error}
            </div>
        );
    }
}