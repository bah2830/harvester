class App extends React.Component {
    render() {
        return (
            <div>
                <Toolbar />
                <Timers />
            </div>
        );
    }
}

if (typeof timers !== 'undefined') {
    const render = () =>  ReactDOM.render(<App />, document.getElementById('app'));
    timers.render = render;
}