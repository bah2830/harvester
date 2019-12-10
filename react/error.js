import React from 'react';

export class Error extends React.Component {
    render() {
        return (
            <div id="error-content" className="alert alert-danger" role="alert">
                {appData.data.error}
            </div>
        );
    }
}