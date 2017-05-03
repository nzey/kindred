import React from 'react';
import ReactDOM from 'react-dom';
import * as d3 from 'd3';
import Faux from 'react-faux-dom';
import * as topojson from 'topojson';
import {bindActionCreators} from 'redux';
import {connect} from 'react-redux';
import '../../styles/index.css';

class DataMap extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      stateAbbr: {
        'AL': 'Alabama',
        'AK': 'Alaska',
        'AZ': 'Arizona',
        'AR': 'Arkansas',
        'CA': 'California',
        'CO': 'Colorado',
        'CT': 'Connecticut',
        'DE': 'Delaware',
        'DC': 'District of Columbia',
        'FL': 'Florida',
        'GA': 'Georgia',
        'HI': 'Hawaii',
        'ID': 'Idaho',
        'IL': 'Illinois',
        'IN': 'Indiana',
        'IA': 'Iowa',
        'KS': 'Kansas',
        'KY': 'Kentucky',
        'LA': 'Louisiana',
        'ME': 'Maine',
        'MD': 'Maryland',
        'MA': 'Massachusetts',
        'MI': 'Michigan',
        'MN': 'Minnesota',
        'MS': 'Mississippi',
        'MO': 'Missouri',
        'MT': 'Montana',
        'NE': 'Nebraska',
        'NV': 'Nevada',
        'NH': 'New Hampshire',
        'NJ': 'New Jersey',
        'NM': 'New Mexico',
        'NY': 'New York',
        'NC': 'North Carolina',
        'ND': 'North Dakota',
        'OH': 'Ohio',
        'OK': 'Oklahoma',
        'OR': 'Oregon',
        'PA': 'Pennsylvania',
        'RI': 'Rhode Island',
        'SC': 'South Carolina',
        'SD': 'South Dakota',
        'TN': 'Tennessee',
        'TX': 'Texas',
        'UT': 'Utah',
        'VT': 'Vermont',
        'VA': 'Virginia',
        'WA': 'Washington',
        'WV': 'West Virginia',
        'WI': 'Wisconsin',
        'WY': 'Wyoming'
      }
    };
    setTimeout(this.sizeChange, 100);
    this.renderMap = this.renderMap.bind(this);
    this.showHoverInfo = this.showHoverInfo.bind(this);
  }

  /* ---- Merge question-of-the-day data with dataless topoJson object ---- */

  mergeTopoWithSelectedStateData(selectedTopic, allStateData, topoData) {
    let selection = selectedTopic ? selectedTopic : allStateData ? Object.keys(allStateData)[0] : '';
    if (allStateData) {
      topoData.objects.usStates.geometries.forEach((topoState, i) => {
        let state = topoState.properties.STATE_ABBR;
        topoData.objects.usStates.geometries[i].properties.data = allStateData[selection][state];
      });
      this.setState({mergeData: topoData});
    }
  }

  componentWillReceiveProps(nextprops) {
    this.mergeTopoWithSelectedStateData(nextprops.questionChoice, nextprops.stateData, nextprops.topoData);
  }
  
  /* ----------------------- Make map size responsive --------------------- */

  sizeChange() {
    d3.select('g')
      .attr('transform', 'scale(' + $('#mapcontainer').width() / 900 + ')');
    $('svg').height($('#mapcontainer').width() * 0.618);
  }

  componentDidMount() {
    d3.select(window).on('resize', this.sizeChange);
  }


  /* ------------------- Build map with data-full topoJson --------------- */

  renderMap(topoData) {
    var datamapContainer = Faux.createElement('div');   
      
    d3.select(datamapContainer)
      .attr('id', 'mapcontainer');

    var svg = d3.select(datamapContainer).append('svg')
      .attr('width', '100%')
        .append('g')
        .classed('no-mouse', true);
    
    var projection = d3.geoAlbersUsa()
      .scale(900);
    
    var path = d3.geoPath()
      .projection(projection);

    const stateSvgs = svg.selectAll('.states')
      .data(topoData)
      .enter()
      .append('path')
      .attr('class', 'states')
      .attr('d', path);

    return datamapContainer;
  }

  /* ------------------------ Build Hovering Info Box --------------------- */
  
  attachHoverBox(domElement) {
    return d3.select(domElement)
      .append('div')
      .attr('id', 'hoverinfo')
      .classed('hide', true);
  }

  showHoverInfo(hoverinfoBox, d) {
    var name = this.state.stateAbbr[d.properties.STATE_ABBR];  
    let text = `Total: ${d.properties.data.total}<br>`;     
    for (let answer in d.properties.data.answers) {
      text += `${answer}: ${d.properties.data.answers[answer]}<br>`;
    }
    return d3.select(hoverinfoBox)
      .classed('hide', false)
      .html(`<strong>${name}</strong><br/>${text}`);
  }

  moveElementWithMouse(element) {
    d3.select(element)
      .style('top', (d3.event.pageY - 10) + 'px')
      .style('left', (d3.event.pageX + 10) + 'px');
  }

  hideElement(element) {
    d3.select(element)
      .classed('hide', true);
  }
  
  populateHoverBox(hoverInfoElement, statesElements) {
    statesElements.on('mouseover', (d) => {
      this.showHoverInfo(hoverinfo, d);
    })
    .on('mousemove', () => {
      this.moveElementWithMouse(hoverinfo);
    })
    .on('mouseout', () => {
      this.hideElement(hoverinfo);
    });
  }

  /* ---------------------------------------------------------------------- */

  render() {
    if (this.state.mergeData) {
      let topoData = topojson.feature(this.state.mergeData, this.state.mergeData.objects.usStates).features;
      let datafullMap = this.renderMap(topoData);
      let hoverBox = this.attachHoverBox(datafullMap);
      let dataElements = d3.select(datafullMap).selectAll('.states')
      this.populateHoverBox(hoverBox, dataElements);
      return datafullMap.toReact();
    } else {
      return null;
    }
  }
}

function mapStateToProps (state) {
  return {
    stateData: state.stateDataReducer,
    topoData: state.topoData,
    questionChoice: state.qotdSelectMap
  };
}

function mapDispatchToProps (dispatch) {
  return {};
}

export default connect(mapStateToProps, mapDispatchToProps)(DataMap);
