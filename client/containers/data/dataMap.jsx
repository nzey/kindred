import React from 'react';
import ReactDOM from 'react-dom';
// import axios from 'axios';
import * as d3 from 'd3';
import topojson from 'topojson';
import Datamap from 'datamaps/dist/datamaps.usa.min'
import objectAssign from 'object-assign';
import {bindActionCreators} from 'redux';
import {connect} from 'react-redux';
import QotdList from './pastQotdsList.jsx';
import '../../styles/index.css';

class DataMap extends React.Component {
  constructor() {
    super();
    this.state = {
      question: 'Questions to look up',
      datamap: null
    };
  }

  linearPalleteScale(value){
    const dataValues = this.props.regionData.map(function(data) { return data.value; });
    const minVal = Math.min(...dataValues);
    const maxVal = Math.max(...dataValues);
    return d3.scaleLinear().domain([minVal, maxVal]).range(["#EFEFFF", "#02386F"])(value);
  }

  reducedData(){
    const newData = this.props.regionData.reduce((object, data) => {
      object[data.code] = { value: data.value, fillColor: this.linearPalleteScale(data.value) };
      return object;
    }, {});
    return objectAssign({}, this.props.stateDefaults, newData);
  }

  renderMap(){
    return new Datamap({
      element: ReactDOM.findDOMNode(this),
      scope: 'usa',
      data: this.reducedData(),
      geographyConfig: {
        highlightBorderColor: '#bada55', 
        highlightBorderWidth: 0.5,
        highlightFillColor: '#FFCC80',
        popupTemplate: function(geography, data) {
          if (data && data.value) {
            console.log("hovering geo name: ", geography.properties.name) ;
            return '<div class="hoverinfo"><strong>' + geography.properties.name + ', ' + data.value + '</strong></div>';
          } else {
            console.log("No data to display");
            return '<div class="hoverinfo"><strong>' + geography.properties.name + '</strong></div>';
          }
        }
      }
    });
  }

  currentScreenWidth(){
    return window.innerWidth ||
        document.documentElement.clientWidth ||
        document.body.clientWidth;
  }

  componentDidMount(){
    const mapContainer = d3.select('#datamap-container');
    const initialScreenWidth = this.currentScreenWidth();
    const containerWidth = (initialScreenWidth < 600) ?
      { width: initialScreenWidth + 'px', height: (initialScreenWidth * 0.5625) + 'px' } :
      { width: '600px', height: '500px' } //'350px'
    mapContainer.style(containerWidth);
    mapContainer.style({overflow: 'inherit'});
    this.state.datamap = this.renderMap();
    d3.select('.datamap').style({overflow: 'overlay'});

    // TODO: documentation suggests this could be a lot simpler - 
    // on 'resize' event, call map.resize():

    //   window.addEventListener('resize', function() {
    //     map.resize();
    // });
    window.addEventListener('resize', () => {
      this.state.datamap.resize();
      const currentScreenWidth = this.currentScreenWidth();
      const mapContainerWidth = mapContainer.style('width');
      if (this.currentScreenWidth() > 600 && mapContainerWidth !== '600px') {
        d3.select('svg').remove();
        mapContainer.style({
          width: '600px',
          height: '350px'
        });
        this.state.datamap = this.renderMap();
      } else if (this.currentScreenWidth() <= 600) {
        d3.select('svg').remove();
        mapContainer.style({
          width: currentScreenWidth + 'px',
          height: (currentScreenWidth * 0.5625) + 'px'
        });
        this.state.datamap = this.renderMap();
      }
    });
  }

  componentDidUpdate(){
    this.state.datamap.updateChoropleth(this.reducedData());
  }

  componentWillUnmount(){
    d3.select('svg').remove();
  }

  render() {
    console.log("props in map component: ", this.props.regionData);
    return (
      <div className="datamap-outer-container">
        <QotdList/>
        <div id="datamap-container"></div>
      </div>
    );
  }
}

function mapStateToProps (state) {
  return {
    regionData: state.stateDataReducer,
    stateDefaults: state.stateDefaults
  };
}

function mapDispatchToProps (dispatch) {
  return {};
  // bindActionCreators({actionUser: actionUser}, dispatch);
}

export default connect(mapStateToProps, mapDispatchToProps)(DataMap);